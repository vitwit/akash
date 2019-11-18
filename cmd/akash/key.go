package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/mattn/go-isatty"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ovrclk/akash/cmd/akash/session"
	"github.com/ovrclk/akash/cmd/common"
	"github.com/ovrclk/akash/errors"
	. "github.com/ovrclk/akash/util"
	"github.com/spf13/cobra"
)

const (
	// BIP44 HD path used to generate HD wallets
	BIP44Path = "44'/118'/0'/0/0"
)

func keyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage keys",
	}
	cmd.AddCommand(keyCreateCommand())
	cmd.AddCommand(keyListCommand())
	cmd.AddCommand(keyShowCommand())
	cmd.AddCommand(keyRecoverCommand())
	cmd.AddCommand(keyRemoveCommand())
	cmd.AddCommand(keyImportCommand())
	return cmd
}

func keyCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <name>",
		Short:   "create a new key locally",
		Long:    "Create a new key and store it locally.\nTo recover a key using the recovery codes generated by the command, see 'akash help key recover'",
		Example: keyCreateHelp,
		RunE:    session.WithSession(session.RequireRootDir(doKeyCreateCommand)),
	}
	session.AddFlagKeyType(cmd, cmd.Flags())
	return cmd
}

func doKeyCreateCommand(ses session.Session, cmd *cobra.Command, args []string) error {
	var name string
	if len(args) == 1 {
		name = args[0]
	}
	name = ses.Mode().Ask().StringVar(name, "Key Name (required): ", true)
	if len(name) == 0 {
		return fmt.Errorf("required argument missing: name")
	}

	kmgr, err := ses.KeyManager()
	if err != nil {
		return err
	}

	info, err := kmgr.Get(name)

	inBuf := bufio.NewReader(cmd.InOrStdin())

	// Check if a key already exists with given name
	if err == nil && len(info.GetPubKey().Address()) != 0 {
		// Confirmation should happen in interactive mode only
		// for other modes(shell and json) it should fail
		if ses.Mode().IsInteractive() {
			res, err := getConfirmation(fmt.Sprintf(
				"Key `%s` already exists. Do you want to override the key anyway?", name), inBuf)

			if err != nil {
				return err
			}

			if res != true { // If user chose to abort
				return errors.NewArgumentError("received no").WithMessage("aborted")
			}
		} else { // Abort key creation
			return errors.NewArgumentError("Key already exists").WithMessage("aborted")
		}
	}

	ktype, err := ses.KeyType()
	if err != nil {
		return err
	}

	password, err := ses.Password()
	if err != nil {
		return err
	}

	types.GetConfig().SetFullFundraiserPath(BIP44Path)
	info, seed, err := kmgr.CreateMnemonic(name, common.DefaultCodec, password, ktype)
	if err != nil {
		return err
	}

	printer := ses.Mode().Printer()
	printer.Log().WithModule("key").Info("key created")
	data := printer.NewSection("Create Key").NewData()
	data.
		WithTag("raw", info).
		Add("Name", name).
		Add("Public Key", X(info.GetPubKey().Address())).
		Add("Recovery Codes", seed)

	printer.Flush()
	notice := "Write these Recovery codes in a safe place. It is the only way to recover your account."
	printer.Log().WithModule("Important").Warn(notice)
	return nil
}

func keyListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all the keys stored locally",
		RunE:  session.WithSession(session.RequireKeyManager(doKeyListCommand)),
	}
}

func doKeyListCommand(s session.Session, cmd *cobra.Command, args []string) error {
	kmgr, _ := s.KeyManager()
	infos, err := kmgr.List()
	if err != nil {
		return err
	}

	printer := s.Mode().Printer()
	data := printer.NewSection("Key List").NewData().AsList().WithTag("raw", infos)
	for _, info := range infos {
		data.
			Add("Name", info.GetName()).
			Add("Public Key Address", X(info.GetPubKey().Address())).
			WithLabel("Public Key Address", "Public Key (Address)")
	}
	return printer.Flush()
}

func keyRecoverCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "recover <name> <recovery-codes>...",
		Short:   "recover a key using recovery codes",
		Long:    "Recover a key using the recovery code generated during key creation and store it locally. For help with creating a key, see 'akash help key create'",
		Example: keyRecoverExample,
		RunE:    session.WithSession(session.RequireKeyManager(doKeyRecoverCommand)),
	}
}

func doKeyRecoverCommand(ses session.Session, cmd *cobra.Command, args []string) error {
	// the first arg is the key name and the rest are mnemonic codes
	var name, seed string

	if len(args) > 2 {
		name, args = args[0], args[1:]
		seed = strings.Join(args, " ")
	}

	name = ses.Mode().Ask().StringVar(name, "Key Name (required): ", true)
	seed = ses.Mode().Ask().StringVar(seed, "Recovery Codes (required): ", true)
	if len(name) == 0 {
		return errors.NewArgumentError("required argument missing: name")
	}
	if len(seed) == 0 {
		return errors.NewArgumentError("seed")
	}

	password, err := ses.Password()
	if err != nil {
		return err
	}
	kmgr, _ := ses.KeyManager()

	params, err := hd.NewParamsFromPath(BIP44Path)
	if err != nil {
		return err
	}

	info, err := kmgr.Derive(name, seed, keys.DefaultBIP39Passphrase, password, *params)
	if err != nil {
		return err
	}

	printer := ses.Mode().Printer()
	printer.Log().Info(fmt.Sprintf("Successfully recovered key, stored locally as '%s'", name))

	printer.NewSection("Recover Key").NewData().AsPane().
		WithTag("raw", info).
		Add("Name", name).
		Add("Public Key Address", X(info.GetPubKey().Address())).
		WithLabel("Public Key Address", "Public Key (Address)")
	return printer.Flush()
}

func keyShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "display a key",
		RunE:  session.WithSession(session.RequireRootDir(doKeyShowCommand)),
	}
	cmd.Flags().Bool("public", false, "display only public key")
	cmd.Flags().Bool("private", false, "display only private key")
	return cmd
}

func doKeyShowCommand(ses session.Session, cmd *cobra.Command, args []string) error {
	var name string
	if len(args) > 0 {
		name = args[0]
	}
	name = ses.Mode().Ask().StringVar(name, "Key Name (required): ", true)

	if len(name) == 0 {
		return errors.NewArgumentError("name")
	}

	kmgr, err := ses.KeyManager()
	if err != nil {
		return err
	}
	info, err := kmgr.Get(name)
	if err != nil {
		return err
	}

	if len(info.GetPubKey().Address()) == 0 {
		return fmt.Errorf("key not found %s", name)
	}

	printer := ses.Mode().Printer()

	if ok, _ := cmd.Flags().GetBool("public"); ok {
		fmt.Println(X(info.GetPubKey().Address()))
		return nil
	}

	if ok, _ := cmd.Flags().GetBool("private"); ok {
		info, err := kmgr.Export(name)
		if err != nil {
			return err
		}
		fmt.Println(info)
		return nil
	}

	if ok, _ := cmd.Flags().GetBool("private"); ok {
		fmt.Println(X(info.GetPubKey().Address()))
		return nil
	}

	data := printer.NewSection("Display Key").NewData()
	data.
		WithTag("raw", info).
		Add("Name", name).
		Add("Public Key Address", X(info.GetPubKey().Address())).
		WithLabel("Public Key Address", "Public Key (Address)")
	return printer.Flush()
}

func keyRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "remove key locally",
		Long:  "remove a key from local.\nTo recover a key using the recovery codes generated by the command, see 'akash help key recover'",
		RunE:  session.WithSession(session.RequireRootDir(doKeyRemoveCommand)),
	}
	session.AddFlagKeyType(cmd, cmd.Flags())
	return cmd
}

func keyImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <name> <path>",
		Short: "import a private key",
		Long:  "import a private key with the name from the given path",
		RunE:  session.WithSession(session.RequireRootDir(doKeyImportCommand)),
	}
	session.AddFlagKeyType(cmd, cmd.Flags())
	return cmd
}

func doKeyImportCommand(ses session.Session, cmd *cobra.Command, args []string) error {
	var name, path string
	if len(args) > 1 {
		name = args[0]
		path = args[1]
	}
	name = ses.Mode().Ask().StringVar(name, "Key Name (required): ", true)
	path = ses.Mode().Ask().StringVar(path, "Path (required): ", true)

	if len(name) == 0 {
		return errors.NewArgumentError("name")
	}

	if len(path) == 0 {
		return errors.NewArgumentError("path")
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	kmgr, err := ses.KeyManager()
	if err != nil {
		return err
	}

	err = kmgr.Import(name, string(b))
	if err != nil {
		return err
	}

	p := ses.Mode().Printer()
	p.Log().WithModule("key").Info("key imported")
	data := p.NewSection("Import Key").NewData()
	data.
		Add("Name", name).
		Add("Path", path)
	return p.Flush()
}

func doKeyRemoveCommand(ses session.Session, cmd *cobra.Command, args []string) error {
	var name string
	if len(args) == 1 {
		name = args[0]
	}
	name = ses.Mode().Ask().StringVar(name, "Key Name (required): ", true)
	if len(name) == 0 {
		return fmt.Errorf("required argument missing: name")
	}
	kmgr, err := ses.KeyManager()
	if err != nil {
		return err
	}

	password, err := ses.Password()
	if err != nil {
		return err
	}
	if err := kmgr.Delete(name, password, true); err != nil {
		return nil
	}

	p := ses.Mode().Printer()
	p.Log().WithModule("key").Info("key removed")
	p.NewSection("Remove Key").NewData().
		Add("Name", name)
	return p.Flush()
}

var (
	keyCreateHelp = `
- Create a key with the name 'greg':

  $ akash key create greg

  Successfully created key for 'greg'
  ===================================

  Public Key:    	f4e03226c054b1adafaa2739bad720c095500a49
  Recovery Codes:	figure share industry canal...
`

	keyRecoverExample = `
- Recover a key with the name 'my-key':

  $ akash key recover my-key today napkin arch \
	 picnic fox case thrive table journey ill  \
	 any enforce awesome desert chapter regret \
	 narrow capable advice skull pipe giraffe \
	 clown outside
`
)

func getConfirmation(prompt string, buf *bufio.Reader) (bool, error) {
	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		fmt.Print(fmt.Sprintf("%s [y/N]: ", prompt))
	}

	response, err := buf.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(response)
	if len(response) == 0 {
		return false, nil
	}

	response = strings.ToLower(response)
	if response[0] == 'y' {
		return true, nil
	}

	return false, nil
}
