package music

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/webp"
)

const (
	cardWidth           = 1000
	cardHeight          = 600
	cardStatusFontSize  = 35
	cardTitleFontSize   = 55
	cardArtistFontSize  = 40
	cardPaddingLeftSize = 35
	coverSize           = 350
	cardTitleMaxLines   = 2
	cardArtistMaxLines  = 1
	progressHeight      = 15
)

var progressWidth = 500.0 - cardPaddingLeftSize

var fonts = map[string]map[string]string{
	"japanese": {
		"title":  "assets/fonts/NotoSansJP-Bold.ttf",
		"artist": "assets/fonts/NotoSansJP-Regular.ttf",
	},
	"chinese": {
		"title":  "assets/fonts/NotoSansSC-Bold.ttf",
		"artist": "assets/fonts/NotoSansSC-Regular.ttf",
	},
	"default": {
		"title":  "assets/fonts/NotoSans-Bold.ttf",
		"artist": "assets/fonts/NotoSans-Regular.ttf",
	},
}

func (m *MusicBot) ShowMusicCard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	buf, err := generateMusicCard()
	if err != nil {
		fmt.Println(err)
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Files: []*discordgo.File{
				{
					Name:        "music_card.png",
					ContentType: "image/png",
					Reader:      buf,
				},
			},
		},
	})
	if err != nil {
		fmt.Println(err)
	}
}

func generateMusicCard() (*bytes.Buffer, error) {
	// var wg sync.WaitGroup
	dc := gg.NewContext(cardWidth, cardHeight)

	randomTitles := []string{
		"混沌を越えし我らが神聖",
		"(9.0) 混沌を越えし我らが神聖なる調律主を讃えよ [MASTER 15+] (譜面確認) [CHUNITHM チュウニズム]",
		"混沌を越えし我らが神聖なる調律主を讃えよ",
		"热爱105°C的你",
		"Takut",
		"Aku Patut Membenci Dia",
		"ごめんね、SUMMER",
	}
	randomArtist := []string{
		"Eufonius",
		"紬 ヴェンダース",
		"阿肆",
		"Hello",
		"ABCDERFJFJAJAJ",
	}
	randomImages := []string{
		"https://images-ext-1.discordapp.net/external/5ShJ-tzlhriLZQb3NhkdXmHVCxCeUBhPKMNIZ4rWBlU/https/i.ytimg.com/vi/y6Vc_troC44/maxresdefault.jpg?format=webp",
		"https://images-ext-1.discordapp.net/external/_RDTMoMjV_-wjJtlnduu8wCNO-_wo8PilmwxnlbBbeA/https/i.scdn.co/image/ab67616d0000b27320dee54944c831048ae1a25c?format=webp",
		"https://images-ext-1.discordapp.net/external/5mK-BimDWD9dhlK_9QZoRLhTHbSd_pipDLf0Ej8uFjI/https/i.scdn.co/image/ab67616d0000b27306562a7ad2a1c7698409c5f5?format=webp&width=80&height=80",
	}
	addedName := "Added By : moonchild30s"
	platformName := "Platform : Spotify"
	queueLengthInfo := "Num. tracks in queue : 100"

	// Get title and artist using random function from golang
	artist := randomArtist[rand.IntN(len(randomArtist))]
	title := randomTitles[rand.IntN(len(randomTitles))]
	albumUrl := randomImages[rand.IntN(len(randomImages))]
	fmt.Println(artist, title, albumUrl)

	// Get fontface path
	artistFontFacePath := fonts[getFontFaces(artist)]["artist"]
	titleFontFacePath := fonts[getFontFaces(title)]["title"]

	// Background
	dc.SetRGB(0.07, 0.07, 0.07)
	dc.Clear()

	appNameFontFace := loadFontFace(fonts["default"]["title"], cardStatusFontSize)
	artistFontFace := loadFontFace(artistFontFacePath, cardArtistFontSize)
	titleFontFace := loadFontFace(titleFontFacePath, cardTitleFontSize)
	durationFontFace := loadFontFace(fonts["default"]["artist"], 28)
	additionalInfoAddedFontFace := loadFontFace(fonts["default"]["artist"], 28)
	additionalInfoFontFace := loadFontFace(fonts["default"]["artist"], 25)

	// Draw Left Side (Contains status, title, artist, and progress bar)
	dc.SetFontFace(appNameFontFace)
	dc.SetRGB(1, 1, 1)
	dc.DrawString("Chisa Music Player (test)", cardPaddingLeftSize, 80)

	dc.SetRGB(1, 1, 1)
	drawWordWrap(dc, title, cardPaddingLeftSize, (cardHeight/2)-25, (cardWidth/2)-cardPaddingLeftSize, cardTitleMaxLines, titleFontFace)

	dc.SetFontFace(artistFontFace)
	dc.SetRGB(1, 1, 1)
	drawWordWrap(dc, artist, cardPaddingLeftSize, (cardHeight/2)+55, (cardWidth/2)-cardPaddingLeftSize, cardArtistMaxLines, artistFontFace)

	// Draw Progressbar
	current := 45.0 // seconds
	total := 180.0  // seconds
	currentDurationText := "00:45"
	totalDurationText := "03:00"

	progress := current / total
	barY := (cardHeight / 2) + 130.0

	dc.SetFontFace(durationFontFace)

	currentDurationWidth, _ := dc.MeasureString(currentDurationText)
	totalDurationWidth, _ := dc.MeasureString(totalDurationText)

	fmt.Println(currentDurationWidth, totalDurationWidth)
	dc.DrawString(currentDurationText, cardPaddingLeftSize, barY+(progressHeight))
	dc.DrawString(totalDurationText, progressWidth-40, barY+(progressHeight))

	progressWidthNow := progressWidth - currentDurationWidth - totalDurationWidth - 40

	dc.SetRGB(0.3, 0.3, 0.3) // background bar
	dc.DrawRoundedRectangle(cardPaddingLeftSize+currentDurationWidth+20, barY, progressWidthNow, progressHeight, 5)
	dc.Fill()

	dc.SetRGB(0.2, 0.7, 0.2) // green progress
	dc.DrawRoundedRectangle(cardPaddingLeftSize+currentDurationWidth+20, barY, progressWidthNow*progress, progressHeight, 5)
	dc.Fill()

	// Add Additional Information (Who added, queue length and provider)
	dc.SetFontFace(additionalInfoAddedFontFace)
	dc.SetRGB(1, 1, 1)
	dc.DrawString(addedName, cardPaddingLeftSize, barY+100)

	dc.SetFontFace(additionalInfoFontFace)
	dc.DrawString(platformName, cardPaddingLeftSize, barY+150)
	w, _ := dc.MeasureString(platformName)
	dc.DrawString(queueLengthInfo, w+50, barY+150)

	// Draw Right Side (Contains album image)
	// Draw album image
	img := loadAlbumImageFromURL(albumUrl)
	if img != nil {
		alX := (cardWidth / 2) + 50
		alY := (cardHeight / 2) - (coverSize / 2)
		cover := cropOrResizeAlbumImage(img)
		dc.DrawImage(cover, alX, alY)
	}

	// Encode to buffer
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}
	return &buf, nil
}

func loadFontFace(path string, size float64) font.Face {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(path)
		panic(err)
	}
	ft, err := opentype.Parse(fontBytes)
	if err != nil {
		panic(err)
	}
	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}
	return face
}

func loadAlbumImageFromURL(url string) image.Image {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Failed to download image:", err)
		return nil
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err == nil {
		return img
	}
	if err.Error() != "image: unknown format" {
		fmt.Println("Failed to decode image:", err)
	}
	img, err = webp.Decode(resp.Body)
	if err != nil {
		fmt.Println("Failed to decode webp image:", err)
		return nil
	}
	return img
}

func cropOrResizeAlbumImage(img image.Image) image.Image {
	// check width and height same
	bounds := img.Bounds()

	if bounds.Dx() > bounds.Dy() {
		focalX := (bounds.Dx() - coverSize) / 2
		focalY := (bounds.Dy() - coverSize) / 2
		croppedImage := image.Rect(focalX, focalY, focalX+coverSize, focalY+coverSize)
		croppedImage = croppedImage.Intersect(img.Bounds())

		res := image.NewRGBA(image.Rect(0, 0, croppedImage.Dx(), croppedImage.Dy()))
		draw.Draw(res, res.Bounds(), img, croppedImage.Min, draw.Src)
		return res
	}
	return resize.Resize(coverSize, coverSize, img, resize.Lanczos3)
}

func getFontFaces(text string) string {
	for _, r := range text {
		switch {
		case unicode.Is(unicode.Han, r):
			return "chinese"
		case unicode.Is(unicode.Hiragana, r), unicode.Is(unicode.Katakana, r):
			return "japanese"
		}
	}
	return "default"
}

func isCJK(r rune) bool {
	return unicode.In(r,
		unicode.Han,
		unicode.Hiragana,
		unicode.Katakana,
		unicode.Hangul,
	)
}

func drawWordWrap(dc *gg.Context, title string, x, y, maxWidth float64, maxLines int, fontFace font.Face) {
	dc.SetFontFace(fontFace)
	words := splitWords(title)
	var lines []string
	var line string

	for _, word := range words {
		if lineWidth, _ := dc.MeasureString(line + word); lineWidth > maxWidth {
			// If last string of line is space, remove it before append to lines
			if line[len(line)-1] == ' ' {
				line = line[:len(line)-1]
			}

			if len(lines) == maxLines-1 {
				lines = append(lines, line[:len(line)-3]+"...")
				break
			}
			lines = append(lines, line)
			line = ""
		}
		line += word
		if !isCJK([]rune(word)[0]) {
			line += " "
		}
	}
	if line != "" && len(lines) < maxLines {
		lines = append(lines, line)
	}

	lineHeight := dc.FontHeight()
	if lineHeight == 0 {
		lineHeight = 80
	}
	totalHeight := float64(len(lines)) * lineHeight

	startY := y - totalHeight
	for i, line := range lines {
		dc.DrawStringAnchored(line, x, startY+float64(i+1)*lineHeight, 0, 0)
	}
}

func splitWords(text string) []string {
	var words []string
	var word strings.Builder
	lastCJK := new(bool)
	flush := func() {
		if word.Len() == 0 {
			return
		}
		if *lastCJK {
			// CJK run → emit each rune separately
			for _, r := range word.String() {
				words = append(words, string(r))
			}
		} else {
			// Non‑CJK run → split on spaces
			fields := strings.Fields(word.String())
			words = append(words, fields...)
		}
		word.Reset()
	}

	for _, r := range text {
		if unicode.IsSpace(r) {
			flush()
			continue
		}
		cjk := isCJK(r)
		if lastCJK == nil {
			*lastCJK = cjk
		} else if *lastCJK != cjk {
			flush()
			*lastCJK = cjk
		}
		word.WriteRune(r)
	}
	flush()
	return words
}
