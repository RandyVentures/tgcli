package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/RandyVentures/tgcli/internal/out"
)

func newSyncCmd(flags *rootFlags) *cobra.Command {
	var follow bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync messages (requires prior auth)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, lk, err := newApp(ctx, flags, true, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			tgClient := a.TGClient()
			if tgClient == nil {
				return fmt.Errorf("telegram client not initialized")
			}

			// Check if authenticated
			authed, err := tgClient.IsAuthed(ctx)
			if err != nil {
				return wrapErr(err, "check auth status")
			}
			if !authed {
				return fmt.Errorf("not authenticated. Run 'tgcli auth' first")
			}

			if !flags.asJSON {
				fmt.Println("📥 Syncing dialogs...")
			}

			// Sync dialogs
			if err := tgClient.SyncDialogs(ctx); err != nil {
				return wrapErr(err, "sync dialogs failed")
			}

			// Sync recent messages for each chat
			chats, err := a.Store().ListChats()
			if err != nil {
				return wrapErr(err, "list chats failed")
			}

			if !flags.asJSON {
				fmt.Printf("📥 Syncing messages for %d chats...\n", len(chats))
			}

			for i, chat := range chats {
				if !flags.asJSON {
					fmt.Printf("  [%d/%d] %s\n", i+1, len(chats), chat.Title)
				}
				if err := tgClient.SyncChatHistory(ctx, chat.ID, 50); err != nil {
					fmt.Fprintf(os.Stderr, "  Warning: failed to sync %s: %v\n", chat.Title, err)
					continue
				}
			}

			if flags.asJSON {
				return out.WriteJSON(os.Stdout, map[string]interface{}{
					"status":      "synced",
					"chats_count": len(chats),
				})
			}

			fmt.Println()
			fmt.Println("✅ Sync complete!")
			fmt.Printf("📊 Synced %d chats\n", len(chats))

			if follow {
				fmt.Println()
				fmt.Println("👂 Continuous sync mode (--follow) not implemented in Phase 1")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&follow, "follow", false, "continuous sync (stay connected) - not implemented in Phase 1")
	return cmd
}
