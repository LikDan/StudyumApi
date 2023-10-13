package main

import (
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	Connect()
}

func Connect() {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", "admin.api.studyum.net", 4186)},
		Auth: clickhouse.Auth{
			Database: "i18n",
			Username: "studyum",
			Password: "0653b347be8a7b399320833980121107a6ed1ff965d828728d86e1c651b1cc3266f67c75508a81fc8735ad3900e7e64bfdde1df825ee40767746ed48c8f2c0187d8dd2bb677b545a6ac0822e3cd373ba",
		},
		Protocol: clickhouse.HTTP,
	})

	rows, err := conn.Query("SELECT key, group, ru_ru, en_us FROM public WHERE group = 'defaults'")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var (
		group, key, ru, en string
	)
	for rows.Next() {
		if err := rows.Scan(&group, &key, &ru, &en); err != nil {
			return
		}
		fmt.Printf("row: group=%s, key=%s, ru=%s, en=%s\n", group, key, ru, en)
	}
}
