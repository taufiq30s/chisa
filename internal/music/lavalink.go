package music

import (
	"context"
	"fmt"
	"time"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/taufiq30s/chisa/utils"
)

var nodeConfigurations []disgolink.NodeConfig = []disgolink.NodeConfig{
	{
		Name:     "ajieblogs_v407",
		Address:  "lava-v4.ajieblogs.eu.org:443",
		Password: "https://dsc.gg/ajidevserver",
		Secure:   true,
	},
	{
		Name:     "ajieblogs_combined",
		Address:  "lava-all.ajieblogs.eu.org:443",
		Password: "https://dsc.gg/ajidevserver",
		Secure:   true,
	},
	{
		Name:     "serenetia_id_v4",
		Address:  "lavalinkv4-id.serenetia.com:443",
		Password: "BatuManaBisa",
		Secure:   true,
	},
	{
		Name:     "serenetia_eu_v4",
		Address:  "lavalinkv4-eu.serenetia.com:443",
		Password: "BatuManaBisa",
		Secure:   true,
	},
}

func NewLavalinkClient(botUserId string) (disgolink.Client, error) {
	utils.InfoLog.Println("Connecting to lavalink node")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := disgolink.New(
		snowflake.MustParse(botUserId),
	)
	for _, nodeConfiguration := range nodeConfigurations {
		node, err := client.AddNode(ctx, nodeConfiguration)
		if err != nil {
			utils.ErrorLog.Println("Failed to connect to lavalink node")
			return nil, fmt.Errorf("failed to connect to lavalink node: %w", err)
		}

		version, err := node.Version(ctx)
		if err != nil {
			utils.ErrorLog.Println("Failed to get lavalink node version")
			return nil, fmt.Errorf("failed to get lavalink node version: %w", err)
		}
		utils.InfoLog.Printf("Connected to lavalink node: %s version: %s\n", node.Config().Name, version)
		fmt.Printf("Connected to lavalink node: %s version: %s\n", node.Config().Name, version)
	}
	return client, nil
}
