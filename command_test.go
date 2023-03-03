package cobrax_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/haijima/cobrax"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := cobrax.NewCommand(viper.New(), afero.NewMemMapFs())
	cmd.Use = "test"

	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.Command)
	assert.Equal(t, "test", cmd.Name())
}

func TestWrap(t *testing.T) {
	cmd := cobrax.Wrap(&cobra.Command{
		Use: "test",
	}, viper.New(), afero.NewMemMapFs())

	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.Command)
	assert.Equal(t, "test", cmd.Name())
}

// TestCommand_Viper tests the Viper() and SetViper().
func TestCommand_Viper(t *testing.T) {
	v := viper.New()
	cmd := cobrax.NewCommand(v, afero.NewMemMapFs())

	assert.Exactly(t, v, cmd.Viper())
}

// TestCommand_Fs tests the Fs() and SetFs().
func TestCommand_Fs(t *testing.T) {
	fs := afero.NewMemMapFs()
	cmd := cobrax.NewCommand(viper.New(), fs)

	assert.Exactly(t, fs, cmd.Fs())
}

// TestCommand_SubCommand tests AddCommand() and Commands()
func TestCommand_SubCommand(t *testing.T) {
	v := viper.New()
	fs := afero.NewMemMapFs()
	root := cobrax.NewCommand(v, fs)
	root.Use = "root"
	foo := cobrax.NewCommand(v, fs)
	foo.Use = "foo"
	bar := cobrax.NewCommand(v, fs)
	bar.Use = "bar"

	root.AddCommand(foo, bar)

	assert.Equal(t, 2, len(root.Commands()))
	assert.Equal(t, "bar", root.Commands()[0].Name())
	assert.Equal(t, "foo", root.Commands()[1].Name())
	assert.Equal(t, 2, len(root.Command.Commands()))
	assert.Equal(t, "bar", root.Command.Commands()[0].Name())
	assert.Equal(t, "foo", root.Command.Commands()[1].Name())
}

func TestCommand_RemoveCommand(t *testing.T) {
	v := viper.New()
	fs := afero.NewMemMapFs()
	root := cobrax.NewCommand(v, fs)
	root.Use = "root"
	foo := cobrax.NewCommand(v, fs)
	foo.Use = "foo"
	bar := cobrax.NewCommand(v, fs)
	bar.Use = "bar"

	root.AddCommand(foo, bar)
	root.RemoveCommand(foo)

	assert.Equal(t, 1, len(root.Commands()))
	assert.Equal(t, "bar", root.Commands()[0].Name())
	assert.Equal(t, 1, len(root.Command.Commands()))
	assert.Equal(t, "bar", root.Command.Commands()[0].Name())
}

func TestCommand_ResetCommands(t *testing.T) {
	v := viper.New()
	fs := afero.NewMemMapFs()
	one := cobrax.NewCommand(v, fs)
	one.Use = "one"
	two := cobrax.NewCommand(v, fs)
	two.Use = "two"
	three := cobrax.NewCommand(v, fs)
	three.Use = "three"
	four1 := cobrax.NewCommand(v, fs)
	four1.Use = "four1"
	four2 := cobrax.NewCommand(v, fs)
	four2.Use = "four2"
	five := cobrax.NewCommand(v, fs)
	five.Use = "five"

	one.AddCommand(two)
	two.AddCommand(three)
	three.AddCommand(four1, four2)
	four1.AddCommand(five)

	three.ResetCommands()

	assert.Equal(t, 1, len(one.Commands()))
	assert.Equal(t, 1, len(one.Command.Commands()))
	assert.Equal(t, 1, len(two.Commands()))         // two is still have there
	assert.Equal(t, 1, len(two.Command.Commands())) // two is still have there
	assert.False(t, three.Command.HasParent())      // but three has no parent
	assert.Equal(t, 0, len(three.Commands()))
	assert.Equal(t, 0, len(three.Command.Commands()))
	assert.Equal(t, 1, len(four1.Commands()))
	assert.Equal(t, 1, len(four1.Command.Commands()))
}

func TestCommand_Root(t *testing.T) {
	v := viper.New()
	fs := afero.NewMemMapFs()
	root := cobrax.NewCommand(v, fs)
	root.Use = "root"
	child := cobrax.NewCommand(v, fs)
	child.Use = "child"
	grandChild := cobrax.NewCommand(v, fs)
	grandChild.Use = "grandChild"

	child.AddCommand(grandChild)
	root.AddCommand(child)

	assert.Equal(t, root, grandChild.Root())
}

func TestCommand_WalkCommands(t *testing.T) {
	v := viper.New()
	fs := afero.NewMemMapFs()
	root := cobrax.NewCommand(v, fs)
	root.Use = "root"
	foo := cobrax.NewCommand(v, fs)
	foo.Use = "foo"
	bar := cobrax.NewCommand(v, fs)
	bar.Use = "bar"
	baz := cobrax.NewCommand(v, fs)
	baz.Use = "baz"
	buzz := cobrax.NewCommand(v, fs)
	buzz.Use = "buzz"
	qux := cobrax.NewCommand(v, fs)
	qux.Use = "qux"

	root.AddCommand(foo, bar)
	foo.AddCommand(baz)
	bar.AddCommand(buzz)
	buzz.AddCommand(qux)

	var names []string
	root.WalkCommands(func(cmd *cobrax.Command) {
		names = append(names, cmd.Name())
	})

	assert.Equal(t, []string{"root", "bar", "buzz", "qux", "foo", "baz"}, names)
}
func TestCommand_Execute(t *testing.T) {
	cmd := cobrax.NewCommand(viper.New(), afero.NewMemMapFs())
	cmd.Use = "test"
	cmd.Run = func(cmd *cobrax.Command, args []string) {
		cmd.PrintOut("test")
	}
	out := new(bytes.Buffer)
	cmd.SetOut(out)

	err := cmd.Execute()

	assert.NoError(t, err)
	assert.Equal(t, "test", out.String())
}

func TestCommand_ExecuteC(t *testing.T) {
	cmd := cobrax.NewCommand(viper.New(), afero.NewMemMapFs())
	cmd.Use = "test"
	cmd.Run = func(cmd *cobrax.Command, args []string) {
		cmd.PrintOut("test")
	}
	out := new(bytes.Buffer)
	cmd.SetOut(out)

	c, err := cmd.ExecuteC()

	assert.NoError(t, err)
	assert.Exactly(t, cmd.Command, c)
	assert.Equal(t, "test", out.String())
}

func TestCommand_ExecuteContext(t *testing.T) {
	cmd := cobrax.NewCommand(viper.New(), afero.NewMemMapFs())
	cmd.Use = "test"
	cmd.Run = func(cmd *cobrax.Command, args []string) {
		cmd.PrintOut("test")
	}
	out := new(bytes.Buffer)
	cmd.SetOut(out)

	err := cmd.ExecuteContext(context.TODO())

	assert.NoError(t, err)
	assert.Equal(t, "test", out.String())
}

func TestCommand_ExecuteContextC(t *testing.T) {
	cmd := cobrax.NewCommand(viper.New(), afero.NewMemMapFs())
	cmd.Use = "test"
	cmd.Run = func(cmd *cobrax.Command, args []string) {
		cmd.PrintOut("test")
	}
	out := new(bytes.Buffer)
	cmd.SetOut(out)

	c, err := cmd.ExecuteContextC(context.TODO())

	assert.NoError(t, err)
	assert.Exactly(t, cmd.Command, c)
	assert.Equal(t, "test", out.String())
}

func TestCommand_PrintOut(t *testing.T) {
	cmd := cobrax.NewCommand(viper.New(), afero.NewMemMapFs())
	cmd.Use = "test"
	out := new(bytes.Buffer)
	cmd.SetOut(out)

	cmd.PrintOutln("foo")
	cmd.PrintOut("bar")
	cmd.PrintOutf("%d%s", 1, "baz")

	assert.Equal(t, "foo\nbar1baz", out.String())
}

func TestCommand_ReadFileOrStdIn(t *testing.T) {
	v := viper.New()
	fs := afero.NewMemMapFs()
	cmd := cobrax.NewCommand(v, fs)
	cmd.Use = "test"
	cmd.PersistentFlags().String("file", "", "file to read")
	_ = cmd.BindPersistentFlags()

	testfileName := "testdata/test.txt"
	_, _ = fs.Create(testfileName)
	_ = afero.WriteFile(fs, testfileName, []byte("foo"), 0644)
	v.Set("file", testfileName)

	stdin := new(bytes.Buffer)
	cmd.SetIn(stdin)
	stdin.WriteString("bar")

	// when file is set, read from file
	f, err := cmd.ReadFileOrStdIn("file")
	defer func() { _ = f.Close() }()

	assert.NoError(t, err)
	content, err := io.ReadAll(f)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), content)

	// if no file is given, read from stdin
	f, err = cmd.ReadFileOrStdIn("dummy")
	defer func() { _ = f.Close() }()

	assert.NoError(t, err)
	content, err = io.ReadAll(f)
	assert.NoError(t, err)
	assert.Equal(t, []byte("bar"), content)
}

func setupFlags(t *testing.T) *cobrax.Command {
	t.Helper()

	v := viper.New()
	fs := afero.NewMemMapFs()
	parent := cobrax.NewCommand(v, fs)
	parent.Use = "parent"
	child := cobrax.NewCommand(v, fs)
	child.Use = "child"
	parent.AddCommand(child)

	parent.PersistentFlags().String("pp", "", "parent persistent flag")
	_ = parent.PersistentFlags().Set("pp", "a")
	parent.LocalFlags().String("pl", "", "parent local flag")
	_ = parent.LocalFlags().Set("pl", "b")
	child.PersistentFlags().String("cp", "", "child persistent flag")
	_ = child.PersistentFlags().Set("cp", "c")
	child.LocalFlags().String("cl", "", "child local flag")
	_ = child.LocalFlags().Set("cl", "d")

	return child
}

func TestCommand_BindEachFlag(t *testing.T) {
	child := setupFlags(t)

	assert.NoError(t, child.BindFlag("pp"))
	assert.Error(t, child.BindFlag("pl"))
	assert.NoError(t, child.BindFlag("cp"))
	//assert.NoError(t, child.BindFlag("cl"))
	assert.Error(t, child.BindLocalFlag("pp"))
	assert.Error(t, child.BindLocalFlag("pl"))
	assert.NoError(t, child.BindLocalFlag("cp"))
	assert.NoError(t, child.BindLocalFlag("cl"))
	assert.Error(t, child.BindPersistentFlag("pp"))
	assert.Error(t, child.BindPersistentFlag("pl"))
	assert.NoError(t, child.BindPersistentFlag("cp"))
	assert.Error(t, child.BindPersistentFlag("cl"))
	assert.Error(t, child.BindLocalNonPersistentFlag("pp"))
	assert.Error(t, child.BindLocalNonPersistentFlag("pl"))
	assert.Error(t, child.BindLocalNonPersistentFlag("cp"))
	assert.NoError(t, child.BindLocalNonPersistentFlag("cl"))
	assert.NoError(t, child.BindInheritedFlag("pp"))
	assert.Error(t, child.BindInheritedFlag("pl"))
	assert.Error(t, child.BindInheritedFlag("cp"))
	assert.Error(t, child.BindInheritedFlag("cl"))
	assert.Error(t, child.BindNonInheritedFlag("pp"))
	assert.Error(t, child.BindNonInheritedFlag("pl"))
	assert.NoError(t, child.BindNonInheritedFlag("cp"))
	assert.NoError(t, child.BindNonInheritedFlag("cl"))
}

func TestCommand_BindFlags(t *testing.T) {
	child := setupFlags(t)
	err := child.BindFlags()

	assert.NoError(t, err)
	assert.Equal(t, "a", child.Viper().GetString("pp"))
	assert.Equal(t, "", child.Viper().GetString("pl"))
	assert.Equal(t, "c", child.Viper().GetString("cp"))
	//assert.Equal(t, "d", child.Viper().GetString("cl"))
}

func TestCommand_BindLocalFlags(t *testing.T) {
	child := setupFlags(t)
	err := child.BindLocalFlags()

	assert.NoError(t, err)
	assert.Equal(t, "", child.Viper().GetString("pp"))
	assert.Equal(t, "", child.Viper().GetString("pl"))
	assert.Equal(t, "c", child.Viper().GetString("cp"))
	assert.Equal(t, "d", child.Viper().GetString("cl"))
}

func TestCommand_BindPersistentFlags(t *testing.T) {
	child := setupFlags(t)
	err := child.BindPersistentFlags()

	assert.NoError(t, err)
	assert.Equal(t, "", child.Viper().GetString("pp"))
	assert.Equal(t, "", child.Viper().GetString("pl"))
	assert.Equal(t, "c", child.Viper().GetString("cp"))
	assert.Equal(t, "", child.Viper().GetString("cl"))
}

func TestCommand_BindLocalNonPersistentFlags(t *testing.T) {
	child := setupFlags(t)
	err := child.BindLocalNonPersistentFlags()

	assert.NoError(t, err)
	assert.Equal(t, "", child.Viper().GetString("pp"))
	assert.Equal(t, "", child.Viper().GetString("pl"))
	assert.Equal(t, "", child.Viper().GetString("cp"))
	assert.Equal(t, "d", child.Viper().GetString("cl"))
}

func TestCommand_BindInheritedFlags(t *testing.T) {
	child := setupFlags(t)
	err := child.BindInheritedFlags()

	assert.NoError(t, err)
	assert.Equal(t, "a", child.Viper().GetString("pp"))
	assert.Equal(t, "", child.Viper().GetString("pl"))
	assert.Equal(t, "", child.Viper().GetString("cp"))
	assert.Equal(t, "", child.Viper().GetString("cl"))
}

func TestCommand_BindNonInheritedFlags(t *testing.T) {
	child := setupFlags(t)
	err := child.BindNonInheritedFlags()

	assert.NoError(t, err)
	assert.Equal(t, "", child.Viper().GetString("pp"))
	assert.Equal(t, "", child.Viper().GetString("pl"))
	assert.Equal(t, "c", child.Viper().GetString("cp"))
	assert.Equal(t, "d", child.Viper().GetString("cl"))
}

func Example() {
	cmd := cobrax.NewCommand(viper.New(), afero.NewOsFs())
	cmd.Use = "example"
	cmd.Short = "Example command"
	cmd.Long = "Example command with some flags"
	cmd.Example = "example --foo a --bar b"
	cmd.Run = func(cmd *cobrax.Command, args []string) {
		fmt.Println("Hello world!")
	}

	cmd.PersistentFlags().String("foo", "bar", "Some flag")
	cmd.PersistentFlags().String("bar", "baz", "Some other flag")

	_ = cmd.Execute()
	// Output:
	// Hello world!
}

func ExampleNewCommand() {
	cmd := cobrax.NewCommand(viper.New(), afero.NewOsFs())
	cmd.Use = "example"
	cmd.Run = func(cmd *cobrax.Command, args []string) {
		fmt.Println("Hello world!")
	}

	_ = cmd.Execute()
	// Output:
	// Hello world!
}

func ExampleWrap() {
	orig := &cobra.Command{
		Use: "example",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello world!")
		},
	}
	cmd := cobrax.Wrap(orig, viper.New(), afero.NewOsFs())

	_ = cmd.Execute()
	// Output:
	// Hello world!
}

func ExampleCommand_OpenOrStdIn() {
	// Create a dummy file for example purposes
	fs := afero.NewMemMapFs()
	_, _ = fs.Create("file.txt")
	_ = afero.WriteFile(fs, "file.txt", []byte("Hello world!"), 0644)

	cmd := cobrax.NewCommand(viper.New(), fs)
	cmd.Use = "example"
	cmd.Run = func(cmd *cobrax.Command, args []string) {
		file, err := cmd.OpenOrStdIn(cmd.Viper().GetString("file"))
		if err != nil {
			panic(err)
		}
		defer file.Close() // file should be closed by the caller

		_, _ = io.Copy(os.Stdout, file)
	}

	cmd.PersistentFlags().String("file", "", "file to read")
	_ = cmd.BindPersistentFlag("file")

	os.Args = []string{"example", "--file", "file.txt"}
	_ = cmd.Execute()
	// Output:
	// Hello world!
}
