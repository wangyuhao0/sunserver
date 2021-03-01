package collect

import (
	"fmt"
	"github.com/duanhf2012/origin/sysmodule/mongomodule"
)


type DBType int
const (
	MainDB DBType = 1
)


func SyncIndex(s *mongomodule.Session, dbType DBType,dbName string) error{
	var err error
	if dbType == MainDB {
		if err = s.EnsureUniqueIndex(dbName, "Account", []string{"PlatId"});err != nil {
			return err
		}

		if err = s.EnsureUniqueIndex(dbName, "UserInfo", []string{"UserId"});err != nil {
			return err
		}

		if err = s.EnsureIndex(dbName, "MailInfo", []string{"SendToUser"});err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("Invalid dbType %d!",dbType)
}

