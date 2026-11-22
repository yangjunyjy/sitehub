package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"sitehub/gateway/console"
	"sitehub/gateway/internal/config"
)

func consoleUserCommand(load func(*cobra.Command) (config.Config, error)) *cobra.Command {
	root := &cobra.Command{Use: "console-user", Short: "Manage gateway console credentials in the configured StateDir"}
	var role string
	set := &cobra.Command{Use: "set USER", Short: "Create or reset a local console account and revoke its previous sessions", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := load(cmd)
		if err != nil {
			return err
		}
		password, err := readConsolePassword(cmd.InOrStdin(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		defer clear(password)
		if err := console.SetAccount(cfg.Console.StateDir, args[0], role, password); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Gateway console account %s has role %s.\n", args[0], role)
		return nil
	}}
	set.Flags().StringVar(&role, "role", "", "viewer, publisher, or admin")
	_ = set.MarkFlagRequired("role")
	root.AddCommand(set, &cobra.Command{Use: "remove USER", Short: "Remove a local console account and revoke its sessions", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := load(cmd)
		if err != nil {
			return err
		}
		return console.RemoveAccount(cfg.Console.StateDir, args[0])
	}})
	return root
}

func readConsolePassword(input io.Reader, output io.Writer) ([]byte, error) {
	if f, ok := input.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		fmt.Fprint(output, "New console password: ")
		value, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(output)
		if err != nil {
			return nil, err
		}
		fmt.Fprint(output, "Confirm password: ")
		confirmation, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(output)
		defer clear(confirmation)
		if err != nil || !bytes.Equal(value, confirmation) {
			clear(value)
			return nil, fmt.Errorf("password confirmation failed")
		}
		return value, nil
	}
	line, err := bufio.NewReader(io.LimitReader(input, 75)).ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	return bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r")), nil
}
