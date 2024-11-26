package system

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/vanilla-os/albius/core/util"
)

func SetTimezone(targetRoot, tz string) error {
	tzPath := targetRoot + "/etc/timezone"

	err := os.WriteFile(tzPath, []byte(tz), 0o644)
	if err != nil {
		return fmt.Errorf("failed to set timezone: %s", err)
	}

	linkZoneinfoCmd := "ln -sf /usr/share/zoneinfo/%s /etc/localtime"
	if targetRoot != "" {
		err = util.RunInChroot(targetRoot, fmt.Sprintf(linkZoneinfoCmd, tz))
	} else {
		err = util.RunCommand(fmt.Sprintf(linkZoneinfoCmd, tz))
	}
	if err != nil {
		return fmt.Errorf("failed to set timezone: %s", err)
	}

	return nil
}

// AddUser creates a new user and adds it to the groups provided
//
// If password is left empty, password login will be disabled.
// If uid and/or gid are -1, they will be ignored.
func AddUser(targetRoot, username, fullname string, groups []string, password string, uid, gid int) error {
	adduserCmd := "useradd --shell /bin/bash %s %s && usermod -c \"%s\" %s"

	extraArgs := ""
	if uid != -1 {
		extraArgs = " --uid " + fmt.Sprint(uid)
	}
	if gid != -1 {
		extraArgs = " --gid " + fmt.Sprint(gid)
	}

	var err error
	if targetRoot != "" {
		err = util.RunInChroot(targetRoot, fmt.Sprintf(adduserCmd, extraArgs, username, fullname, username))
	} else {
		err = util.RunCommand(fmt.Sprintf(adduserCmd, extraArgs, username, fullname, username))
	}
	if err != nil {
		return fmt.Errorf("failed to create user: %s", err)
	}

	if password != "" {
		passwdCmd := "echo \"%s:%s\" | chpasswd"
		if targetRoot != "" {
			err = util.RunInChroot(targetRoot, fmt.Sprintf(passwdCmd, username, password))
		} else {
			err = util.RunCommand(fmt.Sprintf(passwdCmd, username, password))
		}
		if err != nil {
			return fmt.Errorf("failed to set password: %s", err)
		}
	}

	// No groups were specified, we're done here
	if len(groups) == 0 {
		return nil
	}

	addGroupCmd := "usermod -a -G %s %s"
	groupList := strings.Join(groups, ",")
	if targetRoot != "" {
		err = util.RunInChroot(targetRoot, fmt.Sprintf(addGroupCmd, groupList, username))
	} else {
		err = util.RunCommand(fmt.Sprintf(addGroupCmd, groupList, username))
	}
	if err != nil {
		return fmt.Errorf("failed to add groups to user: %s", err)
	}

	return nil
}

func RemovePackages(targetRoot, pkgRemovePath, removeCmd string) error {
	pkgRemoveContent, err := os.ReadFile(pkgRemovePath)
	if err != nil {
		return fmt.Errorf("failed to read package removal file: %s", err)
	}

	pkgList := strings.ReplaceAll(string(pkgRemoveContent), "\n", " ")
	completeCmd := fmt.Sprintf("%s %s", removeCmd, pkgList)
	if targetRoot != "" {
		err = util.RunInChroot(targetRoot, completeCmd)
	} else {
		err = util.RunCommand(completeCmd)
	}
	if err != nil {
		return fmt.Errorf("failed to remove packages: %s", err)
	}

	return nil
}

func ChangeHostname(targetRoot, hostname string) error {
	hostnamePath := targetRoot + "/etc/hostname"
	err := os.WriteFile(hostnamePath, []byte(hostname+"\n"), 0o644)
	if err != nil {
		return fmt.Errorf("failed to change hostname: %s", err)
	}

	hostsContents := `127.0.0.1	localhost
::1		localhost
127.0.1.1	%s.localdomain	%s
`
	hostsPath := targetRoot + "/etc/hosts"
	err = os.WriteFile(hostsPath, []byte(fmt.Sprintf(hostsContents, hostname, hostname)), 0o644)
	if err != nil {
		return fmt.Errorf("failed to change hosts file: %s", err)
	}

	return nil
}

func SetLocale(targetRoot, locale string) error {
	err := util.RunCommand(fmt.Sprintf("grep %s %s/usr/share/i18n/SUPPORTED", locale, targetRoot))
	if err != nil {
		return fmt.Errorf("locale %s is invalid", locale)
	}

	err = util.RunCommand(fmt.Sprintf("sed -i 's/^\\# \\(%s\\)/\\1/' %s/etc/locale.gen", regexp.QuoteMeta(locale), targetRoot))
	if err != nil {
		return fmt.Errorf("failed to set locale: %s", err)
	}

	if targetRoot != "" {
		err = util.RunInChroot(targetRoot, "locale-gen")
	} else {
		err = util.RunCommand("locale-gen")
	}
	if err != nil {
		return fmt.Errorf("failed to set locale: %s", err)
	}

	localeContents := `LANG=__lang__
LC_NUMERIC=__lang__
LC_TIME=__lang__
LC_MONETARY=__lang__
LC_PAPER=__lang__
LC_NAME=__lang__
LC_ADDRESS=__lang__
LC_TELEPHONE=__lang__
LC_MEASUREMENT=__lang__
LC_IDENTIFICATION=__lang__
`
	localePath := targetRoot + "/etc/default/locale"
	err = os.WriteFile(localePath, []byte(strings.ReplaceAll(localeContents, "__lang__", locale)), 0o644)
	if err != nil {
		return fmt.Errorf("failed to set locale: %s", err)
	}

	return nil
}

func Swapon(targetRoot, swapPart string) error {
	swaponCmd := "swapon %s"
	if targetRoot != "" {
		return util.RunInChroot(targetRoot, fmt.Sprintf(swaponCmd, swapPart))
	} else {
		return util.RunCommand(fmt.Sprintf(swaponCmd, swapPart))
	}
}

func SetKeyboardLayout(targetRoot, kbLayout, kbModel, kbVariant string) error {
	keyboardContents := `# KEYBOARD CONFIGURATION FILE
# Consult the keyboard(5) manual page.
XKBMODEL="%s"
XKBLAYOUT="%s"
XKBVARIANT="%s"
BACKSPACE="guess"
`
	keyboardPath := targetRoot + "/etc/default/keyboard"
	err := os.WriteFile(keyboardPath, []byte(fmt.Sprintf(keyboardContents, kbModel, kbLayout, kbVariant)), 0o644)
	if err != nil {
		return fmt.Errorf("failed to set keyboard layout: %s", err)
	}

	if targetRoot != "" {
		err = util.RunInChroot(targetRoot, "setupcon --save-only")
	} else {
		err = util.RunCommand("setupcon --save-only")
	}
	if err != nil {
		return fmt.Errorf("failed to set keyboard layout: %s", err)
	}

	return nil
}
