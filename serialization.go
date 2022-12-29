package gommunication

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

func Serialize[T any](buf io.Writer, data T) (err error) {
	var face = reflect.ValueOf(data)

	for i := 0; i < face.NumField(); i++ {
		var (
			fld = face.Field(i)
			val = fld.Interface()
		)

		var bSize = binary.Size(val)
		if bSize >= 0 {
			// Fixed size type
			if fld.Kind() == reflect.Slice {
				err = binary.Write(buf, endianness, uint32(bSize))
				if err != nil {
					return
				}
			}

			err = binary.Write(buf, endianness, val)
			if err != nil {
				return
			}
		} else {
			// Annoying type
			switch fld.Kind() {
			case reflect.String:
				var size = fld.Len()
				err = binary.Write(buf, endianness, uint16(size))
				if err != nil {
					return
				}
				for j := 0; j < size; j++ {
					err = binary.Write(buf, endianness, fld.Index(j).Interface())
					if err != nil {
						return
					}
				}
			default:
				return errors.New("invalid type " + fld.Type().String())
			}
		}

	}

	return nil
}

func Deserialize[T any](buf io.Reader, data *T) (err error) {
	var face = reflect.Indirect(reflect.ValueOf(data))

	for i := 0; i < face.NumField(); i++ {
		var (
			fld = face.Field(i)
			val = fld.Interface()
		)

		var bSize = binary.Size(val)
		if bSize >= 0 {
			// Fixed size type
			if fld.Kind() == reflect.Slice {
				var tmp uint32
				err = binary.Read(buf, endianness, &tmp)
				if err != nil {
					return
				}
				var size = int(tmp)
				fld.Set(reflect.MakeSlice(fld.Type(), size, size))
			}

			err = binary.Read(buf, endianness, fld.Addr().Interface())
			if err != nil {
				return
			}
		} else {
			// Annoying type
			switch fld.Kind() {
			case reflect.String:
				var tmp uint16
				err = binary.Read(buf, endianness, &tmp)
				if err != nil {
					return
				}
				var size = int(tmp)
				var strBuffer = make([]byte, size)
				for j := 0; j < size; j++ {
					err = binary.Read(buf, endianness, &strBuffer[j])
					if err != nil {
						return
					}
				}
				fld.SetString(string(strBuffer))
			default:
				return errors.New("invalid type " + fld.Type().String())
			}
		}
	}

	return nil
}
