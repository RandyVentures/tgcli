package main

import (
	"fmt"
	"os"

	"github.com/RandyVentures/tgcli/internal/tg"
	"github.com/spf13/cobra"
)

func newDoctorCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check configuration and connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("tgcli doctor")
			fmt.Println("============")
			fmt.Println()

			// Check token
			token := tg.GetToken()
			if token == "" {
				fmt.Println("❌ TGCLI_BOT_TOKEN: not set")
				fmt.Println("   Run: export TGCLI_BOT_TOKEN=\"your_bot_token\"")
				fmt.Println("   Get a token from @BotFather on Telegram")
				return nil
			}
			fmt.Println("✅ TGCLI_BOT_TOKEN: set")

			// Try to connect
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, true)
			if err != nil {
				fmt.Printf("❌ Store: %v\n", err)
				return nil
			}
			defer closeApp(a, lk)

			fmt.Printf("✅ Store: %s\n", a.StoreDir())

			client, err := a.Client()
			if err != nil {
				fmt.Printf("❌ Bot API: %v\n", err)
				return nil
			}

			me, err := client.GetMe()
			if err != nil {
				fmt.Printf("❌ Bot connection: %v\n", err)
				return nil
			}

			fmt.Printf("✅ Bot: @%s (ID: %d)\n", me.UserName, me.ID)

			// Check database stats
			chats, _ := a.Store().ListChats(ctx, 1000)
			fmt.Printf("📊 Stored chats: %d\n", len(chats))

			if flags.asJSON {
				return writeJSON(os.Stdout, map[string]interface{}{
					"token_set":    true,
					"store_dir":    a.StoreDir(),
					"bot_id":       me.ID,
					"bot_username": me.UserName,
					"stored_chats": len(chats),
				})
			}

			fmt.Println()
			fmt.Println("All checks passed! ✅")
			return nil
		},
	}
}
