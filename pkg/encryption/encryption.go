package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
)

type Encryption interface {
	Encode(b []byte) string

	EncryptString(s string) string
	DecryptString(s string) string

	Encrypt(value interface{})
	Decrypt(value interface{})

	MapReflectField(value reflect.Value, mapFunc func(str string) string)
}

type encryption struct {
	cipher cipher.Block
}

func NewEncryption(secret string) Encryption {
	cipher_, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return nil
	}

	return &encryption{cipher: cipher_}
}

func (e *encryption) Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func (e *encryption) EncryptString(s string) string {
	plaintext := []byte(s)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(e.cipher, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext)
}

func (e *encryption) DecryptString(s string) string {
	ciphertext, _ := base64.URLEncoding.DecodeString(s)

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(e.cipher, iv)

	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}

func (e *encryption) MapReflectField(value reflect.Value, mapFunc func(str string) string) {
	if value.Kind() == reflect.String {
		value.SetString(mapFunc(value.String()))
		return
	}

	if value.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < value.NumField(); i++ {
		tField := value.Type().Field(i)
		field := value.Field(i)

		switch field.Kind() {
		case reflect.Struct:
			e.MapReflectField(field, mapFunc)
			break
		case reflect.Slice, reflect.Array:
			for j := 0; j < field.Len(); j++ {
				switch field.Index(j).Kind() {
				case reflect.String:
					_, ok := tField.Tag.Lookup("encryption")
					if !ok {
						break
					}

					e.MapReflectField(field.Index(j), mapFunc)
					break
				case reflect.Struct:
					e.MapReflectField(field.Index(j), mapFunc)
					break
				}
			}
			break
		case reflect.String:
			_, ok := tField.Tag.Lookup("encryption")
			if !ok {
				continue
			}

			fieldValue := field.String()
			encrypted := mapFunc(fieldValue)
			field.SetString(encrypted)
			break
		}
	}
}

func (e *encryption) Encrypt(value interface{}) {
	e.MapReflectField(reflect.ValueOf(value).Elem(), func(str string) string {
		return e.EncryptString(str)
	})
}

func (e *encryption) Decrypt(value interface{}) {
	e.MapReflectField(reflect.ValueOf(value).Elem(), func(str string) string {
		return e.DecryptString(str)
	})
}
