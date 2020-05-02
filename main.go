package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"bufio"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	_ "github.com/Tnze/go-mc/data/lang/en-us"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/yggdrasil"
)

// Клиент через которого можно манипулировать своего бота
var client = bot.NewClient()

// Переменная конфига
var data = Config{}

// Структура конфига
type Config struct {
	Host     string `json "host"`
	Port     int    `json "port"`
	Login    string `json "login"`
	Password string `json "password"`
	Online   bool   `json "online"`
}

// Функция для переподключения в случае ошибок
func tryJoin() {
	for {
		time.Sleep(time.Second * 10)
		err := client.JoinServer(data.Host, data.Port)
		if err != nil {
			log.Println(err)
		} else {
			break
		}
	}
}

func main() {
	// Чтение файла конфигурации для входа на сервер и лицензионный аккаунт
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	// Парсинг
	err = json.Unmarshal(file, &data)
	// Проверка на ошибки
	if err != nil {
		log.Fatal(err)
	}

	if data.Online {
		// Готовый код для авторизации лицензионного аккаунта
		auth, err := yggdrasil.Authenticate(data.Login, data.Password)
		if err != nil {
			log.Fatal(err)
		}
		client.Auth.UUID, client.Name = auth.SelectedProfile()
		client.AsTk = auth.AccessToken()
	} else {
		// Просто применяем для оффлайна ник
		client.Name = data.Login
	}
	// Присоединяемся на сервер
	tryJoin()
	log.Println("Присоединение успешно")

	// Регистрация прослушивателей на методы
	client.Events.GameStart = onGameStart
	client.Events.ChatMsg = onChatMsg
	client.Events.Disconnect = onDisconnect
	client.Events.PluginMessage = onPluginMessage

	// Запускаем в отдельном потоке
	go thread()

	// Реализация консольного ввода
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		if text == "/quit\n" {
			os.Exit(1)
		} else {
			err = client.Chat(text)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// Поток который будет работать отдельно
func thread() {
	for {
		//JoinGame
		err := client.HandleGame()
		if err != nil {
			log.Println(err)
			tryJoin()
		}
	}
}

func onGameStart() error {
	log.Println("Запуск игры. Теперь можете вводить чат")
	return nil //if err isn't nil, HandleGame() will return it.
}

func onChatMsg(c chat.Message, pos byte) error {
	log.Println("Чат:", c.ClearString()) // output chat message without any format code (like color or bold)
	return nil
}

func onDisconnect(c chat.Message) error {
	log.Println("Отключены из сервера:", c)
	tryJoin()
	return nil
}

func onPluginMessage(channel string, data []byte) error {
	switch channel {
	case "minecraft:brand":
		var brand pk.String
		if err := brand.Decode(bytes.NewReader(data)); err != nil {
			return err
		}
		log.Println("Движок:", brand)
	default:
		log.Println("Сообщение плагина", channel, data)
	}
	return nil
}
