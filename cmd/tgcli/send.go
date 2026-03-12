package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/RandyVentures/tgcli/internal/tg"
)

func newSendCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send messages (text, file)",
	}

	cmd.AddCommand(newSendTextCmd(flags))
	cmd.AddCommand(newSendFileCmd(flags))
	cmd.AddCommand(newEditCmd(flags))
	cmd.AddCommand(newDeleteCmd(flags))
	cmd.AddCommand(newForwardCmd(flags))

	return cmd
}

func newSendTextCmd(flags *rootFlags) *cobra.Command {
	var to int64
	var message string
	var replyTo int

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Send text message",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			sent, err := client.SendText(tg.SendTextOptions{
				ChatID:  to,
				Text:    message,
				ReplyTo: replyTo,
			})
			if err != nil {
				return err
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, map[string]interface{}{
					"message_id": sent.MessageID,
					"chat_id":    sent.Chat.ID,
					"text":       sent.Text,
					"date":       sent.Date,
				})
			}

			fmt.Printf("✅ Message sent (ID: %d)\n", sent.MessageID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&to, "to", 0, "recipient chat ID (required)")
	cmd.Flags().StringVar(&message, "message", "", "message text (required)")
	cmd.Flags().IntVar(&replyTo, "reply-to", 0, "reply to message ID (optional)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("message")

	return cmd
}

func newSendFileCmd(flags *rootFlags) *cobra.Command {
	var to int64
	var file string
	var caption string
	var asPhoto bool

	cmd := &cobra.Command{
		Use:   "file",
		Short: "Send file or photo",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			var sent interface{}
			if asPhoto {
				s, err := client.SendPhoto(tg.SendPhotoOptions{
					ChatID:   to,
					FilePath: file,
					Caption:  caption,
				})
				if err != nil {
					return err
				}
				sent = s
			} else {
				s, err := client.SendFile(tg.SendFileOptions{
					ChatID:   to,
					FilePath: file,
					Caption:  caption,
				})
				if err != nil {
					return err
				}
				sent = s
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, sent)
			}

			fmt.Println("✅ File sent successfully")
			return nil
		},
	}

	cmd.Flags().Int64Var(&to, "to", 0, "recipient chat ID (required)")
	cmd.Flags().StringVar(&file, "file", "", "file path (required)")
	cmd.Flags().StringVar(&caption, "caption", "", "file caption (optional)")
	cmd.Flags().BoolVar(&asPhoto, "photo", false, "send as photo instead of document")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newEditCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var messageID int
	var text string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a message",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			edited, err := client.EditMessage(tg.EditMessageOptions{
				ChatID:    chatID,
				MessageID: messageID,
				Text:      text,
			})
			if err != nil {
				return err
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, map[string]interface{}{
					"message_id": edited.MessageID,
					"chat_id":    edited.Chat.ID,
					"text":       edited.Text,
				})
			}

			fmt.Printf("✅ Message edited (ID: %d)\n", edited.MessageID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&messageID, "message-id", 0, "message ID to edit (required)")
	cmd.Flags().StringVar(&text, "text", "", "new message text (required)")
	_ = cmd.MarkFlagRequired("chat")
	_ = cmd.MarkFlagRequired("message-id")
	_ = cmd.MarkFlagRequired("text")

	return cmd
}

func newDeleteCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var messageID int

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a message",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			err = client.DeleteMessage(tg.DeleteMessageOptions{
				ChatID:    chatID,
				MessageID: messageID,
			})
			if err != nil {
				return err
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, map[string]interface{}{
					"deleted":    true,
					"chat_id":    chatID,
					"message_id": messageID,
				})
			}

			fmt.Printf("✅ Message deleted (ID: %d)\n", messageID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&messageID, "message-id", 0, "message ID to delete (required)")
	_ = cmd.MarkFlagRequired("chat")
	_ = cmd.MarkFlagRequired("message-id")

	return cmd
}

func newForwardCmd(flags *rootFlags) *cobra.Command {
	var toChatID int64
	var fromChatID int64
	var messageID int

	cmd := &cobra.Command{
		Use:   "forward",
		Short: "Forward a message",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withTimeout(cmd.Context(), flags)
			defer cancel()

			a, lk, err := newApp(ctx, flags, false, false)
			if err != nil {
				return err
			}
			defer closeApp(a, lk)

			client, err := a.Client()
			if err != nil {
				return err
			}

			fwd, err := client.ForwardMessage(tg.ForwardMessageOptions{
				ToChatID:   toChatID,
				FromChatID: fromChatID,
				MessageID:  messageID,
			})
			if err != nil {
				return err
			}

			if flags.asJSON {
				return writeJSON(os.Stdout, map[string]interface{}{
					"message_id":   fwd.MessageID,
					"chat_id":      fwd.Chat.ID,
					"forward_date": fwd.ForwardDate,
				})
			}

			fmt.Printf("✅ Message forwarded (new ID: %d)\n", fwd.MessageID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&toChatID, "to", 0, "destination chat ID (required)")
	cmd.Flags().Int64Var(&fromChatID, "from", 0, "source chat ID (required)")
	cmd.Flags().IntVar(&messageID, "message-id", 0, "message ID to forward (required)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("message-id")

	return cmd
}
