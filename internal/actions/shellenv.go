package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
)

var (
	Bash        = "bash"         // shell selector
	Csh         = "csh"          // shell selector
	Fish        = "fish"         // shell selector
	Pwsh        = "pwsh"         // shell selector
	PwshPreview = "pwsh-preview" // shell selector
	Sh          = "sh"           // shell selector
	Tcsh        = "tcsh"         // shell selector
	Zsh         = "zsh"          // shell selector

	// Shells lists all supported shells.
	Shells = []string{
		Bash,
		Csh,
		Fish,
		Pwsh,
		Sh,
		Tcsh,
		Zsh,
	}
)

// ShellEnv represents the action and its options.
type ShellEnv struct {
	*Hops

	Shell string // selects the shell type
}

// Run runs the action.
func (action *ShellEnv) Run(_ context.Context) error {
	brew := action.Homebrew()

	if action.Shell == "" {
		shell, ok := os.LookupEnv("SHELL")
		if !ok {
			// This will yell at the user every time the output is evaluated
			fmt.Println(`echo "ERROR(hops shellenv): SHELL is not set"`)
			os.Exit(1)
		}
		// fmt.Printf(`echo "INFO(hops shellenv): SHELL is set to %s"`+"\n", shell)
		action.Shell = filepath.Base(shell)
	}

	brewRepository := brew.Prefix
	bin := filepath.Join(brew.Prefix, "bin")
	sbin := filepath.Join(brew.Prefix, "bin")
	manpath := filepath.Join(brew.Prefix, "share", "man")
	infopath := filepath.Join(brew.Prefix, "share", "info")
	backtick := "`"

	// Check if environment is already modified.
	// Check if the PATH starts with "HOMEBREW_PREFIX/bin:HOMEBREW_PREFIX/sbin"
	// Or if the entire PATH is "HOMEBREW_PREFIX/bin"
	// I didn't make these rules, they are from the brew shellenv script.
	//
	// Bash conditional:
	// [[ "${HOMEBREW_PATH%%:"${HOMEBREW_PREFIX}"/sbin*}" == "${HOMEBREW_PREFIX}/bin" ]]
	path := os.Getenv("PATH")
	if path == bin || strings.HasPrefix(path, bin+string(os.PathListSeparator)+sbin) {
		return nil
	}

	switch strings.TrimPrefix(action.Shell, "-") {
	case Fish:
		/*
			fish | -fish)
				echo "set -gx HOMEBREW_PREFIX \"${HOMEBREW_PREFIX}\";"
				echo "set -gx HOMEBREW_CELLAR \"${HOMEBREW_CELLAR}\";"
				echo "set -gx HOMEBREW_REPOSITORY \"${HOMEBREW_REPOSITORY}\";"
				echo "fish_add_path -gP \"${HOMEBREW_PREFIX}/bin\" \"${HOMEBREW_PREFIX}/sbin\";"
				echo "! set -q MANPATH; and set MANPATH ''; set -gx MANPATH \"${HOMEBREW_PREFIX}/share/man\" \$MANPATH;"
				echo "! set -q INFOPATH; and set INFOPATH ''; set -gx INFOPATH \"${HOMEBREW_PREFIX}/share/info\" \$INFOPATH;"
				;;
		*/
		fmt.Println(heredoc.Docf(`
			set -gx HOMEBREW_PREFIX %q;
			set -gx HOMEBREW_CELLAR %q;
			set -gx HOMEBREW_REPOSITORY %q;
			fish_add_path -gP %q %q;
			! set -q MANPATH; and set MANPATH ''; set -gx MANPATH %q $MANPATH;
			! set -q INFOPATH; and set INFOPATH ''; set -gx INFOPATH %q $INFOPATH;`,
			brew.Prefix,
			action.Prefix().Cellar(),
			brewRepository,
			bin, sbin,
			manpath,
			infopath,
		))
	case Csh, Tcsh:
		/*
			csh | -csh | tcsh | -tcsh)
				echo "setenv HOMEBREW_PREFIX ${HOMEBREW_PREFIX};"
				echo "setenv HOMEBREW_CELLAR ${HOMEBREW_CELLAR};"
				echo "setenv HOMEBREW_REPOSITORY ${HOMEBREW_REPOSITORY};"
				echo "setenv PATH ${HOMEBREW_PREFIX}/bin:${HOMEBREW_PREFIX}/sbin:\$PATH;"
				echo "setenv MANPATH ${HOMEBREW_PREFIX}/share/man\`[ \${?MANPATH} == 1 ] && echo \":\${MANPATH}\"\`:;"
				echo "setenv INFOPATH ${HOMEBREW_PREFIX}/share/info\`[ \${?INFOPATH} == 1 ] && echo \":\${INFOPATH}\"\`;"
				;;
		*/
		fmt.Println(heredoc.Docf(`
			setenv HOMEBREW_PREFIX %s;
			setenv HOMEBREW_CELLAR %s;
			setenv HOMEBREW_REPOSITORY %s;
			setenv PATH %s:%s:$PATH;
			setenv MANPATH %s%s[ ${?MANPATH} == 1 ] && echo ":${MANPATH}"%s:;
			setenv INFOPATH %s%s[ ${?INFOPATH} == 1 ] && echo ":${INFOPATH}"%s;`,
			action.Prefix(),
			action.Prefix().Cellar(),
			brewRepository,
			bin, sbin,
			manpath, backtick, backtick,
			infopath, backtick, backtick,
		))
	case Pwsh, PwshPreview:
		/*
			pwsh | -pwsh | pwsh-preview | -pwsh-preview)
				echo "[System.Environment]::SetEnvironmentVariable('HOMEBREW_PREFIX','${HOMEBREW_PREFIX}',[System.EnvironmentVariableTarget]::Process)"
				echo "[System.Environment]::SetEnvironmentVariable('HOMEBREW_CELLAR','${HOMEBREW_CELLAR}',[System.EnvironmentVariableTarget]::Process)"
				echo "[System.Environment]::SetEnvironmentVariable('HOMEBREW_REPOSITORY','${HOMEBREW_REPOSITORY}',[System.EnvironmentVariableTarget]::Process)"
				echo "[System.Environment]::SetEnvironmentVariable('PATH',\$('${HOMEBREW_PREFIX}/bin:${HOMEBREW_PREFIX}/sbin:'+\$ENV:PATH),[System.EnvironmentVariableTarget]::Process)"
				echo "[System.Environment]::SetEnvironmentVariable('MANPATH',\$('${HOMEBREW_PREFIX}/share/man'+\$(if(\${ENV:MANPATH}){':'+\${ENV:MANPATH}})+':'),[System.EnvironmentVariableTarget]::Process)"
				echo "[System.Environment]::SetEnvironmentVariable('INFOPATH',\$('${HOMEBREW_PREFIX}/share/info'+\$(if(\${ENV:INFOPATH}){':'+\${ENV:INFOPATH}})),[System.EnvironmentVariableTarget]::Process)"
				;;
		*/
		fmt.Println(heredoc.Docf(`
			[System.Environment]::SetEnvironmentVariable('HOMEBREW_PREFIX','%s',[System.EnvironmentVariableTarget]::Process);
			[System.Environment]::SetEnvironmentVariable('HOMEBREW_CELLAR','%s',[System.EnvironmentVariableTarget]::Process);
			[System.Environment]::SetEnvironmentVariable('HOMEBREW_REPOSITORY','%s',[System.EnvironmentVariableTarget]::Process);
			[System.Environment]::SetEnvironmentVariable('PATH',$('%s:%s:'+$ENV:PATH),[System.EnvironmentVariableTarget]::Process);
			[System.Environment]::SetEnvironmentVariable('MANPATH',$('%s'+$(if(${ENV:MANPATH}){':'+${ENV:MANPATH}})+':'),[System.EnvironmentVariableTarget]::Process)
			[System.Environment]::SetEnvironmentVariable('INFOPATH',$('%s'+$(if(${ENV:INFOPATH}){':'+${ENV:INFOPATH}})),[System.EnvironmentVariableTarget]::Process)`,
			action.Prefix(),
			action.Prefix().Cellar(),
			brewRepository,
			bin, sbin,
			manpath,
			infopath,
		))
	default:
		/*
			*)
				echo "export HOMEBREW_PREFIX=\"${HOMEBREW_PREFIX}\";"
				echo "export HOMEBREW_CELLAR=\"${HOMEBREW_CELLAR}\";"
				echo "export HOMEBREW_REPOSITORY=\"${HOMEBREW_REPOSITORY}\";"
				echo "export PATH=\"${HOMEBREW_PREFIX}/bin:${HOMEBREW_PREFIX}/sbin\${PATH+:\$PATH}\";"
				echo "export MANPATH=\"${HOMEBREW_PREFIX}/share/man\${MANPATH+:\$MANPATH}:\";"
				echo "export INFOPATH=\"${HOMEBREW_PREFIX}/share/info:\${INFOPATH:-}\";"
				;;
		*/
		fmt.Println(heredoc.Docf(`
			export HOMEBREW_PREFIX=%q;
			export HOMEBREW_CELLAR=%q;
			export HOMEBREW_REPOSITORY=%q;
			export PATH="%s:%s${PATH+:$PATH}";
			export MANPATH="%s${MANPATH+:$MANPATH}:";
			export INFOPATH="%s:${INFOPATH:-}";`,
			action.Prefix(),
			action.Prefix().Cellar(),
			brewRepository,
			bin, sbin,
			manpath,
			infopath,
		))
	}

	return nil
}
