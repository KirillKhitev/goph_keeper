package mycrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

// Encrypt шифрует тело запроса используя ключ пользователя.
func Encrypt(src []byte, keyFile string) ([]byte, error) {
	result := []byte{}

	if keyFile == "" {
		return result, fmt.Errorf("не передано название файла с ключом")
	}

	keyBase64, err := os.ReadFile("users" + string(os.PathSeparator) + keyFile)

	if err != nil {
		return result, fmt.Errorf("error by read key file, error: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(string(keyBase64))

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("error cipher: %v\n", err)
		return result, fmt.Errorf("ошибка при создании блока для шифрования: %w", err)
	}

	// NewGCM возвращает заданный 128-битный блочный шифр
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Printf("error NewGCM: %v\n", err)
		return result, err
	}

	// создаём вектор инициализации
	nonce := key[len(key)-aesgcm.NonceSize():]
	if err != nil {
		fmt.Printf("error nonce: %v\n", err)
		return result, err
	}

	dst := aesgcm.Seal(nil, nonce, src, nil) // зашифровываем

	return dst, nil
}

// Decrypt расшифровывает тело запроса используя ключ пользователя.
func Decrypt(data []byte, keyFile string) ([]byte, error) {
	result := []byte{}

	if keyFile == "" {
		return result, fmt.Errorf("не передано название файла с ключом")
	}

	keyBase64, err := os.ReadFile("users" + string(os.PathSeparator) + keyFile)

	if err != nil {
		return result, fmt.Errorf("error by read key file, error: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(string(keyBase64))

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("error cipher: %v\n", err)
		return result, fmt.Errorf("ошибка при создании блока для шифрования: %w", err)
	}

	// NewGCM возвращает заданный 128-битный блочный шифр
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		log.Printf("error NewGCM: %v\n", err)
		return result, err
	}

	// создаём вектор инициализации
	nonce := key[len(key)-aesgcm.NonceSize():]
	if err != nil {
		log.Printf("error nonce: %v\n", err)
		return result, err
	}

	res, err := aesgcm.Open(nil, nonce, data, nil) // расшифровываем
	if err != nil {
		log.Printf("error: %v\n", err)
		return result, err
	}

	return res, err
}

func GenerateRandom(size int) ([]byte, error) {
	// генерируем криптостойкие случайные байты в b
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
