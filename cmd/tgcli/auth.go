package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/RandyVentures/tgcli/internal/out"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	var follow bool

	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Telegram (phone + code)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, true, true)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			tgClient := a.TGClient()
			if tgClient == nil {
				return fmt.Errorf("telegram client not initialized (missing TGCLI_APP_ID or TGCLI_APP_HASH?)")
			}

			fmt.Println("🔐 Starting authentication...")
			fmt.Println()

			// Perform authentication
			if err := tgClient.Auth(ctx, ""); err != nil {
				return wrapErr(err, "authentication failed")
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]interface{}{
					"status": "authenticated",
				})
			}

			fmt.Println()
			fmt.Println("✅ Authentication successful!")
			fmt.Println()
			fmt.Println("📥 Starting initial sync...")

			// Sync dialogs
			if err := tgClient.SyncDialogs(ctx); err != nil {
				return wrapErr(err, "sync dialogs failed")
			}

			// Sync recent messages for each chat (limit 20 per chat)
			chats, err := a.Store().ListChats()
			if err != nil {
				return wrapErr(err, "list chats failed")
			}

			for i, chat := range chats {
				fmt.Printf("  [%d/%d] Syncing %s...\n", i+1, len(chats), chat.Title)
				if err := tgClient.SyncChatHistory(ctx, chat.ID, 20); err != nil {
					fmt.Fprintf(os.Stderr, "  Warning: failed to sync %s: %v\n", chat.Title, err)
					continue
				}
			}

			fmt.Println()
			fmt.Println("✅ Initial sync complete!")
			fmt.Printf("📊 Synced %d chats\n", len(chats))
			fmt.Println()
			fmt.Println("You can now use 'tgcli sync' to update messages.")

			if follow {
				fmt.Println()
				fmt.Println("👂 Continuous sync mode not implemented in Phase 1")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&follow, "follow", false, "continuous sync after auth (not implemented in Phase 1)")
	return cmd
}
