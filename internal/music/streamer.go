package music

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/ogg"
	"github.com/taufiq30s/chisa/internal/model"
	"github.com/taufiq30s/chisa/internal/music/providers"
)

type Streamer struct {
	voiceConnection *discordgo.VoiceConnection
	proxy           string

	// procMu guards ytdlp/ffmpeg/paused. Pause and Resume are called from
	// interaction goroutines while Play runs on the player goroutine, so these
	// fields need a lock of their own.
	procMu sync.Mutex
	ytdlp  *os.Process
	ffmpeg *os.Process
	paused bool
}

func newStreamer(voiceConnection *discordgo.VoiceConnection, proxy string) *Streamer {
	return &Streamer{
		voiceConnection: voiceConnection,
		proxy:           proxy,
	}
}

// Pause freezes both child processes with SIGSTOP.
//
// Why this works: yt-dlp and ffmpeg are separate OS processes, so we do not
// have to tear down the stream to pause it - we just stop scheduling them. The
// 16K --buffer-size pipe fills up, yt-dlp blocks on write, ffmpeg blocks on
// read, and Streamer.Play blocks inside decoder.Decode(). No packets reach
// OpusSend, so Discord simply hears nothing; the voice connection stays up.
//
// The trade-off: a very long pause can make YouTube drop the socket. yt-dlp
// then retries (--retries infinite) on resume, so the track may restart.
func (s *Streamer) Pause() error {
	s.procMu.Lock()
	defer s.procMu.Unlock()

	if s.ytdlp == nil || s.ffmpeg == nil {
		return fmt.Errorf("nothing is playing")
	}
	if s.paused {
		return nil
	}
	if err := s.ytdlp.Signal(syscall.SIGSTOP); err != nil {
		return fmt.Errorf("pause yt-dlp: %w", err)
	}
	if err := s.ffmpeg.Signal(syscall.SIGSTOP); err != nil {
		return fmt.Errorf("pause ffmpeg: %w", err)
	}
	s.paused = true
	return nil
}

// Resume undoes Pause with SIGCONT.
func (s *Streamer) Resume() error {
	s.procMu.Lock()
	defer s.procMu.Unlock()

	if s.ytdlp == nil || s.ffmpeg == nil {
		return fmt.Errorf("nothing is playing")
	}
	if !s.paused {
		return nil
	}
	if err := s.ffmpeg.Signal(syscall.SIGCONT); err != nil {
		return fmt.Errorf("resume ffmpeg: %w", err)
	}
	if err := s.ytdlp.Signal(syscall.SIGCONT); err != nil {
		return fmt.Errorf("resume yt-dlp: %w", err)
	}
	s.paused = false
	return nil
}

// IsPaused reports whether playback is currently frozen.
func (s *Streamer) IsPaused() bool {
	s.procMu.Lock()
	defer s.procMu.Unlock()
	return s.paused
}

func (s *Streamer) Play(ctx context.Context, track model.Track) error {
	playCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	ytdlpBinary := providers.FindYtdlpBinary()

	ytdlpArgs := []string{
		"--no-playlist",
		"-f", "bestaudio[abr<=128]/bestaudio",
		"-o", "-",
		"--quiet",
		// resilience for slow/flaky connections
		"--retries", "infinite",
		"--fragment-retries", "infinite",
		"--retry-sleep", "3",
		"--socket-timeout", "15",
		"--extractor-retries", "5",
		"--no-check-certificates",
		"--buffer-size", "16K",
	}
	if s.proxy != "" {
		ytdlpArgs = append(ytdlpArgs, "--proxy", s.proxy)
	}
	ytdlpArgs = append(ytdlpArgs, track.Url)

	ytdlp := exec.CommandContext(playCtx, ytdlpBinary, ytdlpArgs...)
	var ytdlpStderr bytes.Buffer
	ytdlp.Stderr = &ytdlpStderr

	ffmpeg := exec.CommandContext(playCtx, "ffmpeg",
		"-loglevel", "quiet",
		"-i", "pipe:0",
		"-c:a", "libopus",
		"-b:a", "96k",
		"-vbr", "on",
		"-compression_level", "10",
		"-frame_duration", "20",
		"-application", "audio",
		"-f", "ogg",
		"pipe:1",
	)
	var ffmpegStderr bytes.Buffer
	ffmpeg.Stderr = &ffmpegStderr

	ytdlpStdout, err := ytdlp.StdoutPipe()
	if err != nil {
		return fmt.Errorf("yt-dlp pipe: %w", err)
	}
	ffmpeg.Stdin = ytdlpStdout

	ffmpegStdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg pipe: %w", err)
	}

	if err := ytdlp.Start(); err != nil {
		return fmt.Errorf("yt-dlp start: %w", err)
	}
	if err := ffmpeg.Start(); err != nil {
		_ = ytdlp.Process.Kill()
		_ = ytdlp.Wait()
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	// Publish the PIDs so Pause/Resume can signal them, and make sure a new
	// track always starts unpaused even if the previous one was paused.
	s.procMu.Lock()
	s.ytdlp = ytdlp.Process
	s.ffmpeg = ffmpeg.Process
	s.paused = false
	s.procMu.Unlock()

	defer func() {
		s.procMu.Lock()
		s.ytdlp = nil
		s.ffmpeg = nil
		s.paused = false
		s.procMu.Unlock()
	}()

	err = s.voiceConnection.Speaking(true)
	if err != nil {
		return fmt.Errorf("voice connection speaking: %w", err)
	}

	defer func() { _ = s.voiceConnection.Speaking(false) }()

	decoder := ogg.NewPacketDecoder(ogg.NewDecoder(ffmpegStdout))
	headerSkipped := 0
	cancelled := false

	for {
		pkt, _, err := decoder.Decode()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip/Stop kill the children, so Decode fails with EPIPE rather
			// than EOF. That is a normal cancellation, not a playback failure -
			// reporting it would spam the channel with an error on every skip.
			if playCtx.Err() != nil {
				break
			}
			return fmt.Errorf("decode ogg packet: %w", err)
		}

		if headerSkipped < 2 {
			headerSkipped++
			continue
		}

		select {
		case <-playCtx.Done():
			// Do NOT return here. Skipping the Waits below leaves yt-dlp and
			// ffmpeg as unreaped zombie processes - one pair per skip/stop.
			// Break out so the Wait calls still run.
			cancelled = true
		case s.voiceConnection.OpusSend <- pkt:
		}
		if cancelled {
			break
		}
	}

	if err := ytdlp.Wait(); err != nil && playCtx.Err() == nil {
		return fmt.Errorf("yt-dlp exit: %w, stderr: %s", err, ytdlpStderr.String())
	}
	if err := ffmpeg.Wait(); err != nil && playCtx.Err() == nil {
		return fmt.Errorf("ffmpeg exit: %w, stderr: %s", err, ffmpegStderr.String())
	}

	return nil
}

// There is deliberately no Stop() method here. Cancelling playback is the
// player loop's job: it owns the context passed to Play (see
// MusicBot.stopPlayback). A Streamer-level cancel would be a second, unsynced
// way to stop the same thing, and the field it used was written by the player
// goroutine while being read by interaction goroutines.
