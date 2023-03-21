package albius

import (
	"fmt"
	"os"
)

// Set locale
// Set keyboard

func SetTimezone(targetRoot, tz string) error {
	tzPath := targetRoot + "/etc/timezone"

	err := os.WriteFile(tzPath, []byte(tz), 0644)
	if err != nil {
		return fmt.Errorf("Failed to set timezone: %s", err)
	}

	return nil
}

func AddUser(targetRoot, username, fullname string, groups []string, withPassword bool, password ...string) error {
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
