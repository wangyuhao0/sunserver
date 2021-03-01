package util

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const RemarksRowsCount = 5
const ServerSkipCount = 3
var Comma = ','
var Comment = '#'

type FileItem struct {
	KindStr string
	IndexOf int
}

func generateFileItemMap(nameLine, typeLine []string, numField int, mapFileItem map[string]FileItem) {
	if len(nameLine) < numField || len(typeLine) < numField || len(nameLine) != len(typeLine) {
		return
	}

	for i := 0; i < len(nameLine); i++ {
		if nameLine[i] == "" || typeLine[i] == "" {
			continue
		}

		name := nameLine[i]
		fileItem := FileItem{
			KindStr: typeLine[i],
			IndexOf: i,
		}
		mapFileItem[name] = fileItem
	}

	return
}

type Index map[interface{}]interface{}
type RecordFile struct {
	Comma      rune
	Comment    rune
	typeRecord reflect.Type
	records    []interface{}
	indexes    []Index
}

func NewRecordFile(st interface{}) (*RecordFile, error) {
	typeRecord := reflect.TypeOf(st)
	if typeRecord == nil || typeRecord.Kind() != reflect.Struct {
		return nil, errors.New("st must be a struct")
	}

	for i := 0; i < typeRecord.NumField(); i++ {
		f := typeRecord.Field(i)

		kind := f.Type.Kind()
		switch kind {
		case reflect.Bool:
		case reflect.Int:
		case reflect.Int8:
		case reflect.Int16:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Uint:
		case reflect.Uint8:
		case reflect.Uint16:
		case reflect.Uint32:
		case reflect.Uint64:
		case reflect.Float32:
		case reflect.Float64:
		case reflect.String:
		case reflect.Struct:
		case reflect.Array:
		case reflect.Slice:
		case reflect.Map:
		default:
			return nil, fmt.Errorf("invalid type: %v %s", f.Name, kind)
		}

		tag := f.Tag
		if tag == "index" {
			switch kind {
			case reflect.Slice, reflect.Map, reflect.Float32, reflect.Float64:
				return nil, fmt.Errorf("could not index %s field %v %v", kind, i, f.Name)
			}
		}
	}

	rf := new(RecordFile)
	rf.typeRecord = typeRecord

	return rf, nil
}

func (rf *RecordFile) Read(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	if rf.Comma == 0 {
		rf.Comma = Comma
	}
	if rf.Comment == 0 {
		rf.Comment = Comment
	}
	reader := csv.NewReader(file)
	reader.Comma = rf.Comma
	reader.Comment = rf.Comment
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}

	typeRecord := rf.typeRecord

	// make records
	records := make([]interface{}, len(lines) - RemarksRowsCount)

	// make indexes
	indexes := []Index{}
	for i := 0; i < typeRecord.NumField(); i++ {
		tag := typeRecord.Field(i).Tag
		if tag == "index" {
			indexes = append(indexes, make(Index))
		}
	}

	n := ServerSkipCount
	//第四行开始为服务器标记位
	//第四行为数据名、第五行为数据类型
	nameLine := lines[n]
	typeLine := lines[n+1]
	mapFileItem := make(map[string]FileItem, typeRecord.NumField())
	generateFileItemMap(nameLine, typeLine, typeRecord.NumField(), mapFileItem)

	n += 2
	recordsIndex := 0
	for ; n < len(lines); n++ {
		value := reflect.New(typeRecord)
		records[recordsIndex] = value.Interface()
		record := value.Elem()

		line := lines[n]
		if len(line) < typeRecord.NumField() {
			return fmt.Errorf("line %v, field count mismatch: %v (file) %v (st)", n, len(line), typeRecord.NumField())
		}

		iIndex := 0

		for i := 0; i < typeRecord.NumField(); i++ {
			f := typeRecord.Field(i)

			itemKind := f.Type.Kind()
			itemTypeStr := f.Type.Kind().String()
			fileItem, ok := mapFileItem[f.Name]
			if !ok || itemTypeStr != fileItem.KindStr || fileItem.IndexOf >= len(line) {
				return fmt.Errorf("parse field (row=%v, col=%v) error: %v", n, i, err)
			}

			// records
			strField := line[fileItem.IndexOf]
			field := record.Field(i)
			if !field.CanSet() {
				continue
			}

			var err error
			if itemKind == reflect.Bool {
				var v bool
				v, err = strconv.ParseBool(strField)
				if err == nil {
					field.SetBool(v)
				}
			} else if itemKind == reflect.Int ||
				itemKind == reflect.Int8 ||
				itemKind == reflect.Int16 ||
				itemKind == reflect.Int32 ||
				itemKind == reflect.Int64 {
				strField = rf.checkKeyWords(strField)
				var v int64
				v, err = strconv.ParseInt(strField, 0, f.Type.Bits())
				if err == nil {
					field.SetInt(v)
				}
			} else if itemKind == reflect.Uint ||
				itemKind == reflect.Uint8 ||
				itemKind == reflect.Uint16 ||
				itemKind == reflect.Uint32 ||
				itemKind == reflect.Uint64 {
				strField = rf.checkKeyWords(strField)
				var v uint64
				v, err = strconv.ParseUint(strField, 0, f.Type.Bits())
				if err == nil {
					field.SetUint(v)
				}
			} else if itemKind == reflect.Float32 ||
				itemKind == reflect.Float64 {
				var v float64
				v, err = strconv.ParseFloat(strField, f.Type.Bits())
				if err == nil {
					field.SetFloat(v)
				}
			} else if itemKind == reflect.String {
				field.SetString(strField)
			} else if itemKind == reflect.Struct ||
				itemKind == reflect.Array ||
				itemKind == reflect.Slice ||
				itemKind == reflect.Map {
				err = json.Unmarshal([]byte(strField), field.Addr().Interface())
			}

			if err != nil {
				return fmt.Errorf("data[%s] parse field (row=%v, col=%v) error: %+v", strField, n, i, err)
			}

			// indexes
			if f.Tag == "index" {
				index := indexes[iIndex]
				iIndex++
				if _, ok := index[field.Interface()]; ok {
					return fmt.Errorf("index error: duplicate at (row=%v, col=%v)", n, i)
				}
				index[field.Interface()] = records[recordsIndex]
			}
		}

		recordsIndex++
	}

	rf.records = records
	rf.indexes = indexes

	return nil
}

func (rf *RecordFile) Record(i int) interface{} {
	return rf.records[i]
}

func (rf *RecordFile) NumRecord() int {
	return len(rf.records)
}

func (rf *RecordFile) Indexes(i int) Index {
	if i >= len(rf.indexes) {
		return nil
	}
	return rf.indexes[i]
}

func (rf *RecordFile) Index(i interface{}) interface{} {
	index := rf.Indexes(0)
	if index == nil {
		return nil
	}
	return index[i]
}

func (rf *RecordFile) checkKeyWords(value string) string {
	vStr := strings.Split(value, "_")
	if len(vStr) >= 2 {
		return vStr[0]
	}

	return value
}
