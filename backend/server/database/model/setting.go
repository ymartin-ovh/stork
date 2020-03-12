package dbmodel

import (
	"strconv"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"
)

const SettingValTypeInt = 1
const SettingValTypeBool = 2
const SettingValTypeStr = 3
const SettingValTypePasswd = 4

// Represents a setting held in setting table in the database.
type Setting struct {
	Name    string `pg:",pk"`
	ValType int64
	Value   string
}

// Initialize settings in db. If new setting needs to be added then add it to defaultSettings list
// and it will be automatically added to db here in this function.
func InitializeSettings(db *pg.DB) error {
	// list of all stork settings with default values
	defaultSettings := []Setting{{
		Name:    "kea_stats_puller_interval", // in seconds
		ValType: SettingValTypeInt,
		Value:   "60",
	}}

	// get present settings from db
	var settings []Setting
	q := db.Model(&settings)
	err := q.Select()
	if err != nil {
		err = errors.Wrapf(err, "problem with getting settings from db")
		return err
	}

	// Check if there are new settings vs existing ones. Add new ones to DB.
	for _, sDef := range defaultSettings {
		// check if setting already exist, if so then skip it
		found := false
		for _, s := range settings {
			if sDef.Name == s.Name {
				found = true
				break
			}
		}
		if found {
			continue
		}

		// if setting is not yet in db, then add it with default value
		sDefTmp := sDef
		err := db.Insert(&sDefTmp)
		if err != nil {
			err = errors.Wrapf(err, "problem with inserting setting %s", sDef.Name)
			return err
		}
	}
	return nil
}

// Get setting record from db based on its name.
func GetSetting(db *pg.DB, name string) (*Setting, error) {
	setting := Setting{}
	q := db.Model(&setting).Where("setting.name = ?", name)
	err := q.Select()
	if err == pg.ErrNoRows {
		return nil, errors.Wrapf(err, "setting %s is missing", name)
	} else if err != nil {
		return nil, errors.Wrapf(err, "problem with getting setting %s", name)
	}
	return &setting, nil
}

// Get int value of given setting by name.
func GetSettingInt(db *pg.DB, name string) (int64, error) {
	s, err := GetSetting(db, name)
	if err != nil {
		return 0, err
	}
	if s.ValType != SettingValTypeInt {
		return 0, errors.Errorf("not matching setting type of %s (%d vs %d expected)", name, s.ValType, SettingValTypeInt)
	}
	val, err := strconv.ParseInt(s.Value, 10, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// Get bool value of given setting by name.
func GetSettingBool(db *pg.DB, name string) (bool, error) {
	s, err := GetSetting(db, name)
	if err != nil {
		return false, err
	}
	if s.ValType != SettingValTypeBool {
		return false, errors.Errorf("not matching setting type of %s (%d vs %d expected)", name, s.ValType, SettingValTypeBool)
	}
	val, err := strconv.ParseBool(s.Value)
	if err != nil {
		return false, err
	}
	return val, nil
}

// Get string value of given setting by name.
func GetSettingStr(db *pg.DB, name string) (string, error) {
	s, err := GetSetting(db, name)
	if err != nil {
		return "", err
	}
	if s.ValType != SettingValTypeStr {
		return "", errors.Errorf("not matching setting type of %s (%d vs %d expected)", name, s.ValType, SettingValTypeStr)
	}
	return s.Value, nil
}

// Get password value of given setting by name.
func GetSettingPasswd(db *pg.DB, name string) (string, error) {
	s, err := GetSetting(db, name)
	if err != nil {
		return "", err
	}
	if s.ValType != SettingValTypePasswd {
		return "", errors.Errorf("not matching setting type of %s (%d vs %d expected)", name, s.ValType, SettingValTypePasswd)
	}
	return s.Value, nil
}

// Set int value of given setting by name.
func SetSettingInt(db *pg.DB, name string, value int64) error {
	s, err := GetSetting(db, name)
	if err != nil {
		return err
	}
	if s.ValType != SettingValTypeInt {
		return errors.Errorf("not matching setting type of %s (%d vs %d expected)", name, s.ValType, SettingValTypeInt)
	}
	s.Value = strconv.FormatInt(value, 10)
	err = db.Update(s)
	if err != nil {
		return errors.Wrapf(err, "problem with updating setting %s", name)
	}
	return nil
}

// Set bool value of given setting by name.
func SetSettingBool(db *pg.DB, name string, value bool) error {
	s, err := GetSetting(db, name)
	if err != nil {
		return err
	}
	if s.ValType != SettingValTypeBool {
		return errors.Errorf("not matching setting type of %s (%d vs %d expected)", name, s.ValType, SettingValTypeBool)
	}
	s.Value = strconv.FormatBool(value)
	err = db.Update(s)
	if err != nil {
		return errors.Wrapf(err, "problem with updating setting %s", name)
	}
	return nil
}

// Set string value of given setting by name.
func SetSettingStr(db *pg.DB, name string, value string) error {
	s, err := GetSetting(db, name)
	if err != nil {
		return err
	}
	if s.ValType != SettingValTypeStr {
		return errors.Errorf("not matching setting type of %s (%d vs %d expected)", name, s.ValType, SettingValTypeStr)
	}
	s.Value = value
	err = db.Update(s)
	if err != nil {
		return errors.Wrapf(err, "problem with updating setting %s", name)
	}
	return nil
}

// Set password value of given setting by name.
func SetSettingPasswd(db *pg.DB, name string, value string) error {
	s, err := GetSetting(db, name)
	if err != nil {
		return err
	}
	if s.ValType != SettingValTypePasswd {
		return errors.Errorf("not matching setting type of %s (%d vs %d expected)", name, s.ValType, SettingValTypePasswd)
	}
	s.Value = value
	err = db.Update(s)
	if err != nil {
		return errors.Wrapf(err, "problem with updating setting %s", name)
	}
	return nil
}