package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long: `To load completion run

. <(metalctl completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(metalctl completion)
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.GenBashCompletion(os.Stdout)
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var zshCompletionCmd = &cobra.Command{
	Use:   "zsh-completion",
	Short: "Generates Z shell completion scripts",
	Long: `To load completion run

. <(metalctl zsh-completion)

To configure your Z shell (with oh-my-zshell framework) to load completions for each session run

echo -e '#compdef _metalctl metalctl\n. <(metalctl zsh-completion)' > $ZSH/completions/_metalctl
rm -f ~/.zcompdump*
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.GenZshCompletion(os.Stdout)
		if err != nil {
			log.Fatalln(err)
		}

	},
}

const bashCompletionFunc = `
__metalctl_get_images()
{
    local template
    template="{{ .id }}"
    local metalctl_out
    if metalctl_out=$(metalctl image list -o template --template="${template}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${metalctl_out}[*]" -- "$cur" ) )
    fi
}

__metalctl_get_partitions()
{
    local template
    template="{{ .id }}"
    local metalctl_out
    if metalctl_out=$(metalctl partition list -o template --template="${template}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${metalctl_out}[*]" -- "$cur" ) )
    fi
}

__metalctl_get_sizes()
{
    local template
    template="{{ .id }}"
    local metalctl_out
    if metalctl_out=$(metalctl size list -o template --template="${template}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${metalctl_out}[*]" -- "$cur" ) )
    fi
}

__metalctl_get_machines()
{
    local template
    template="{{ .id }}"
    local metalctl_out
    if metalctl_out=$(metalctl machine list -o template --template="${template}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${metalctl_out}[*]" -- "$cur" ) )
    fi
}

__metalctl_get_networks()
{
    local template
    template="{{ .id }}"
    local metalctl_out
    if metalctl_out=$(metalctl network list -o template --template="${template}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${metalctl_out}[*]" -- "$cur" ) )
    fi
}

__metalctl_get_ips()
{
    local template
    template="{{ .ipaddress }}"
    local metalctl_out
    if metalctl_out=$(metalctl network ip list -o template --template="${template}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${metalctl_out}[*]" -- "$cur" ) )
    fi
}

__metalctl_get_projects()
{
    local template
    template="{{ if .allocation }} {{ .allocation.project }} {{ end }}"
    local metalctl_out
    if metalctl_out=$(metalctl machine list -o template --template="${template}" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${metalctl_out}[*]" -- "$cur" ) )
    fi
}

__metalctl_get_output_formats()
{
    COMPREPLY=( $( compgen -W "table wide markdown json yaml template" -- "$cur" ) )
}

__metalctl_get_orders()
{
    COMPREPLY=( $( compgen -W "size id status event when partition project" -- "$cur" ) )
}
`
