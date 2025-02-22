package containers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/containers/common/pkg/completion"
	"github.com/containers/podman/v4/cmd/podman/common"
	"github.com/containers/podman/v4/cmd/podman/registry"
	"github.com/containers/podman/v4/pkg/domain/entities"
	"github.com/spf13/cobra"
)

var (
	commitDescription = `Create an image from a container's changes. Optionally tag the image created, set the author with the --author flag, set the commit message with the --message flag, and make changes to the instructions with the --change flag.`

	commitCommand = &cobra.Command{
		Use:               "commit [options] CONTAINER [IMAGE]",
		Short:             "Create new image based on the changed container",
		Long:              commitDescription,
		RunE:              commit,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: common.AutocompleteCommitCommand,
		Example: `podman commit -q --message "committing container to image" reverent_golick image-committed
  podman commit -q --author "firstName lastName" reverent_golick image-committed
  podman commit -q --pause=false containerID image-committed
  podman commit containerID`,
	}

	containerCommitCommand = &cobra.Command{
		Args:              commitCommand.Args,
		Use:               commitCommand.Use,
		Short:             commitCommand.Short,
		Long:              commitCommand.Long,
		RunE:              commitCommand.RunE,
		ValidArgsFunction: commitCommand.ValidArgsFunction,
		Example: `podman container commit -q --message "committing container to image" reverent_golick image-committed
  podman container commit -q --author "firstName lastName" reverent_golick image-committed
  podman container commit -q --pause=false containerID image-committed
  podman container commit containerID`,
	}
)

var (
	commitOptions = entities.CommitOptions{
		ImageName: "",
	}
	iidFile string
)

func commitFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	changeFlagName := "change"
	flags.StringArrayVarP(&commitOptions.Changes, changeFlagName, "c", []string{}, "Apply the following possible instructions to the created image (default []): "+strings.Join(common.ChangeCmds, " | "))
	_ = cmd.RegisterFlagCompletionFunc(changeFlagName, common.AutocompleteChangeInstructions)

	formatFlagName := "format"
	flags.StringVarP(&commitOptions.Format, formatFlagName, "f", "oci", "`Format` of the image manifest and metadata")
	_ = cmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteImageFormat)

	iidFileFlagName := "iidfile"
	flags.StringVarP(&iidFile, iidFileFlagName, "", "", "`file` to write the image ID to")
	_ = cmd.RegisterFlagCompletionFunc(iidFileFlagName, completion.AutocompleteDefault)

	messageFlagName := "message"
	flags.StringVarP(&commitOptions.Message, messageFlagName, "m", "", "Set commit message for imported image")
	_ = cmd.RegisterFlagCompletionFunc(messageFlagName, completion.AutocompleteNone)

	authorFlagName := "author"
	flags.StringVarP(&commitOptions.Author, authorFlagName, "a", "", "Set the author for the image committed")
	_ = cmd.RegisterFlagCompletionFunc(authorFlagName, completion.AutocompleteNone)

	flags.BoolVarP(&commitOptions.Pause, "pause", "p", false, "Pause container during commit")
	flags.BoolVarP(&commitOptions.Quiet, "quiet", "q", false, "Suppress output")
	flags.BoolVarP(&commitOptions.Squash, "squash", "s", false, "squash newly built layers into a single new layer")
	flags.BoolVar(&commitOptions.IncludeVolumes, "include-volumes", false, "Include container volumes as image volumes")
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: commitCommand,
	})
	commitFlags(commitCommand)

	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: containerCommitCommand,
		Parent:  containerCmd,
	})
	commitFlags(containerCommitCommand)
}

func commit(cmd *cobra.Command, args []string) error {
	container := args[0]
	if len(args) == 2 {
		commitOptions.ImageName = args[1]
	}
	if !commitOptions.Quiet {
		commitOptions.Writer = os.Stderr
	}

	response, err := registry.ContainerEngine().ContainerCommit(context.Background(), container, commitOptions)
	if err != nil {
		return err
	}
	if len(iidFile) > 0 {
		if err = ioutil.WriteFile(iidFile, []byte(response.Id), 0644); err != nil {
			return fmt.Errorf("failed to write image ID: %w", err)
		}
	}
	fmt.Println(response.Id)
	return nil
}
