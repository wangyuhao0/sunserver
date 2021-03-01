package command

import (
	_ "github.com/duanhf2012/origin/sysmodule/mongomodule"
	_ "sunserver/gamecore/playerservice/dbcollection"
)

func init() {
	//console.RegisterCommandString("syncindex", "", "Synchironize database index.", SyncIndex)
}

//-syncindex mongodb://admin:123456@127.0.0.1:27017/dbname
/*
func SyncIndex(args interface{}) error {
	url := args.(string)
	if url =="" {
		return nil
	}
	module := mongomodule.MongoModule{}
	index := strings.LastIndex(url, "/")
	input := bufio.NewScanner(os.Stdin)

	connUrl := url[:index]
	dbName := url[index+1:]
	fmt.Printf("The following databases will be synchronized\nURL:%s\nDatabaseName:%s\n", connUrl, dbName)
	fmt.Println("Are you sure want to start?[Y/N]")

	err := module.Init(connUrl, 1, 5*time.Second, 5*time.Second)
	if err != nil {
		fmt.Println("Connection database fail:%s", err.Error())
		return nil
	}
	input.Scan()

	if input.Text() != "Y" {
		return nil
	}

	s := module.Take()
	collect.SyncIndex(s, dbName)

	fmt.Println(" Synchronization index complet")
	return nil
}
*/