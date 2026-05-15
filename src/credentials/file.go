package credentials

import (
	"errors"
	"os"
	"zipcode/src/config"
)

func atomicWrite(contents []byte) error {
	tmp, err := os.CreateTemp(config.Cfg.HomeDir, "credentials-temp-*.toml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}

	if err := os.Rename(tmpName, config.Cfg.CredentialsPath); err != nil {
		os.Remove(tmpName)
		return err
	}

	return nil
}

func enforcePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode()
	if mode.Perm() > 0600 {
		return errors.New("INVALID_PERMS")
	}

	return nil
}

func ensureDir() error {
	_, err := os.Stat(config.Cfg.HomeDir)
	if errors.Is(err, os.ErrNotExist) {
		os.Mkdir(config.Cfg.HomeDir, 0600)
		return nil
	}
	return err
}
