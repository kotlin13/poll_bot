package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserFlowCaseSensitive(t *testing.T) {
	s, err := newTestStorage()
	require.NoError(t, err)
	err = s.init()
	require.NoError(t, err)

	p := generateOurPoll()

	userId := "123"

	reply, err := generateReplyFor(p, s, newSubscribeCallback(t, userId))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Добрый день, Vasya. Добро пожаловать")
	text := newTextCallback(t, userId, "Привет")
	require.Equal(t, text.User.Id, userId)

	reply, err = generateReplyFor(p, s, text)
	require.NoError(t, err)
	require.Equal(t, reply.text, "Укажите, пожалуйста, Ваше гражданство?")
	require.Equal(t, reply.options, []string{"Беларусь", "Россия", "Украина", "Казахстан", "Другая страна"})

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "беларусь"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Укажите, пожалуйста, Ваш возраст")
	require.Equal(t, reply.options, []string{"18-24", "25-34", "35-44", "45-54", "55+"})
}

func TestUserFlow(t *testing.T) {
	s, err := newTestStorage()
	require.NoError(t, err)
	err = s.init()
	require.NoError(t, err)

	p := generateOurPoll()

	userId := "123"

	reply, err := generateReplyFor(p, s, newSubscribeCallback(t, userId))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Добрый день, Vasya. Добро пожаловать")

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "Привет"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Укажите, пожалуйста, Ваше гражданство?")
	require.Equal(t, reply.options, []string{"Беларусь", "Россия", "Украина", "Казахстан", "Другая страна"})

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "Россия"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Только граждание Беларуси могут принимать участие! Укажите, пожалуйста, Ваше гражданство?")
	require.Equal(t, reply.options, []string{"Беларусь", "Россия", "Украина", "Казахстан", "Другая страна"})

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "Беларусь"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Укажите, пожалуйста, Ваш возраст")

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "16"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Вам должно быть 18 или больше. Укажите, пожалуйста, Ваш возраст")

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "39"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Примете ли Вы участие в предстоящих выборах Президента?")
	require.Equal(t, reply.options, []string{"Да, приму обязательно", "Да, скорее приму", "Нет, скорее не приму", "Нет, не приму", "Затрудняюсь ответить"})

	user, err := s.fromPersisted(userId)
	require.NoError(t, err)

	require.Equal(t, user.Id, userId)
	require.Equal(t, user.Age, "39")
	require.Equal(t, user.Level, 3)

	seenCallback := newSeenCallback(t, userId)
	require.Equal(t, seenCallback.User.Id, userId)
	reply, err = generateReplyFor(p, s, seenCallback)
	require.NoError(t, err)
	require.Nil(t, reply)

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "Да, приму обязательно"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Когда Вы планируете голосовать?")
	require.Equal(t, reply.options, []string{"Досрочно", "В основной день", "Затрудняюсь ответить"})

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "Досрочно"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "За кого Вы планируете проголосовать?")
	require.Equal(t, reply.options, []string{"Александр Лукашенко", "Сергей Черечень", "Анна Канопацкая", "Андрей Дмитриев", "Светлана Тихановская", "Против всех", "Затрудняюсь ответить"})

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "Александр Лукашенко"))
	require.NoError(t, err)

	user, err = s.fromPersisted(userId)
	require.NoError(t, err)

	require.Equal(t, user.Id, userId)
	require.Equal(t, user.Age, "39")
	require.Equal(t, user.Level, 6)
	require.Equal(t, user.Candidate, "александр лукашенко")

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "Передумал"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Вы уже проголосовали за Александр Лукашенко")

	reply, err = generateReplyFor(p, s, newTextCallback(t, userId, "Передумал"))
	require.NoError(t, err)
	require.Equal(t, reply.text, "Вы уже проголосовали за Александр Лукашенко")

	subscribe := newSubscribeCallback(t, userId)
	user, err = s.Obtain(userId)
	require.NoError(t, err)
	reply, err = generateReplyFor(p, s, subscribe)
	require.NoError(t, err)
	require.Equal(t, reply, "")

	reply, err = generateReplyFor(p, s, newSeenCallback(t, userId))
	require.NoError(t, err)
	require.Equal(t, reply, "")
}

func newSubscribeCallback(t *testing.T, id string) *ViberCallback {
	c := &ViberCallback{
		Event: "subscribed",
		User: User{
			Id:   id,
			Name: "Vasya",
		},
	}

	b, err := json.Marshal(c)
	require.NoError(t, err)

	ret, err := parseCallback(b)
	require.NoError(t, err)

	return ret
}

func newTextCallback(t *testing.T, id string, text string) *ViberCallback {
	json := `{"event":"message","sender":{"id":"%s","Name":"Vasya"},"message":{"type":"text","text":"%s"}}`

	validJson := fmt.Sprintf(json, id, text)

	ret, err := parseCallback([]byte(validJson))
	require.NoError(t, err)

	return ret
}

func newSeenCallback(t *testing.T, id string) *ViberCallback {
	json := `{"event":"seen","user_id":"%s"}`

	validJson := fmt.Sprintf(json, id)

	ret, err := parseCallback([]byte(validJson))
	require.NoError(t, err)

	return ret
}
