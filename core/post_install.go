package albius

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func SetTimezone(targetRoot, tz string) error {
	tzPath := targetRoot + "/etc/timezone\n"

	err := os.WriteFile(tzPath, []byte(tz), 0644)
	if err != nil {
		return fmt.Errorf("Failed to set timezone: %s", err)
	}

	linkZoneinfoCmd := "ln -sf /usr/share/zoneinfo/%s /etc/localtime"
	if targetRoot != "" {
		err = RunInChroot(targetRoot, fmt.Sprintf(linkZoneinfoCmd, tz))
	} else {
		err = RunCommand(fmt.Sprintf(linkZoneinfoCmd, tz))
	}
	if err != nil {
		return fmt.Errorf("Failed to set timezone: %s", err)
	}

	return nil
}

func AddUser(targetRoot, username, fullname string, groups []string, withPassword bool, password ...string) error {
	// TODO: "adduser" isn't distro agnostic. Change to "useradd"?
	adduserCmd := "adduser --quiet --disabled-password --shell /bin/bash --gecos \"%s\" %s"

	var err error
	if targetRoot != "" {
		err = RunInChroot(targetRoot, fmt.Sprintf(adduserCmd, fullname, username))
	} else {
		err = RunCommand(fmt.Sprintf(adduserCmd, fullname, username))
	}
	if err != nil {
		return fmt.Errorf("Failed to create user: %s", err)
	}

	if withPassword {
		passwdCmd := "echo \"%s:%s\" | chpasswd"
		if len(password) < 1 {
			return fmt.Errorf("Password was not provided")
		}

		if targetRoot != "" {
			err = RunInChroot(targetRoot, fmt.Sprintf(passwdCmd, username, password[0]))
		} else {
			err = RunCommand(fmt.Sprintf(passwdCmd, username, password[0]))
		}
		if err != nil {
			return fmt.Errorf("Failed to set password: %s", err)
		}
	}

	if len(groups) == 0 {
		return nil
	}
	addGroupCmd := "usermod -a -G %s %s"
	groupList := ""
	for i, group := range groups {
		groupList += group
		if i < len(groups)-1 {
			groupList += ","
		}
	}

	if targetRoot != "" {
		err = RunInChroot(targetRoot, fmt.Sprintf(addGroupCmd, groupList, username))
	} else {
		err = RunCommand(fmt.Sprintf(addGroupCmd, groupList, username))
	}
	if err != nil {
		return fmt.Errorf("Failed to set password: %s", err)
	}

	return nil
}

func RemovePackages(targetRoot, pkgRemovePath, removeCmd string) error {
	pkgRemoveContent, err := os.ReadFile(pkgRemovePath)
	if err != nil {
		return fmt.Errorf("Failed to read package removal file: %s", err)
	}

	pkgList := strings.Replace(string(pkgRemoveContent), "\n", " ", -1)

	completeCmd := fmt.Sprintf("%s %s", removeCmd, pkgList)
	if targetRoot != "" {
		return RunInChroot(targetRoot, completeCmd)
	} else {
		return RunCommand(completeCmd)
	}
}

func ChangeHostname(targetRoot, hostname string) error {
	replaceHostnameCmd := "echo %s > %s/etc/hostname"
	cmd := exec.Command("sh", "-c", fmt.Sprintf(replaceHostnameCmd, hostname, targetRoot))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to change hostname: %s", err)
	}

	hostsContents := `127.0.0.1	localhost
::1		localhost
127.0.1.1	%s.localdomain	%s
`
	hostsPath := targetRoot + "/etc/hosts"
	err = os.WriteFile(hostsPath, []byte(fmt.Sprintf(hostsContents, hostname, hostname)), 0644)
	if err != nil {
		return fmt.Errorf("Failed to change hosts file: %s", err)
	}

	return nil
}

func SetLocale(targetRoot, locale string) error {
	err := RunCommand(fmt.Sprintf("grep %s %s/usr/share/i18n/SUPPORTED", locale, targetRoot))
	if err != nil {
		return fmt.Errorf("Locale %s is invalid", locale)
	}

	err = RunCommand(fmt.Sprintf("sed -i 's/^\\# \\(%s\\)/\\1/' %s/etc/locale.gen", regexp.QuoteMeta(locale), targetRoot))
	if err != nil {
		return fmt.Errorf("Failed to set locale: %s", err)
	}

	localeGenCmd := "locale-gen"
	if targetRoot != "" {
		err = RunInChroot(targetRoot, localeGenCmd)
	} else {
		err = RunCommand(localeGenCmd)
	}
	if err != nil {
		return fmt.Errorf("Failed to set locale: %s", err)
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
	err = os.WriteFile(localePath, []byte(strings.ReplaceAll(localeContents, "__lang__", locale)), 0644)
	if err != nil {
		return fmt.Errorf("Failed to set locale: %s", err)
	}

	return nil
}

func Swapon(targetRoot, swapPart string) error {
	swaponCmd := "swapon %s"

	if targetRoot != "" {
		return RunInChroot(targetRoot, fmt.Sprintf(swaponCmd, swapPart))
	} else {
		return RunCommand(fmt.Sprintf(swaponCmd, swapPart))
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
	err := os.WriteFile(keyboardPath, []byte(fmt.Sprintf(keyboardContents, kbModel, kbLayout, kbVariant)), 0644)
	if err != nil {
		return fmt.Errorf("Failed to set keyboard layout: %s", err)
	}

	if targetRoot != "" {
		err = RunInChroot(targetRoot, "setupcon --save-only")
	} else {
		err = RunCommand("setupcon --save-only")
	}
	if err != nil {
		return fmt.Errorf("Failed to set keyboard layout: %s", err)
	}

	return nil
}
