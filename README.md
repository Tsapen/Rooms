
# Система обмена сообщениями  
Данный прототип системы, позволяет общаться нескольким людям через специальные CLI (command line interface) приложения. Для работы реализованы две подсистемы:  
  
Подсистема #1 — сервер, который позволяющий создавать "комнаты" (room) для обмена текстовыми сообщениями, длина каждого сообщения не должна превышает 254 байта, кол-во сообщений в room не превышает 128 — история “Комнаты”. Истории переписок хранятся в течение работы сервера и удаляются при его выключении.  
  
Подсистема #2 — "клиент", который позволяет "публиковать" (publish) сообщения в определенные "комнаты" и  подписываться (subscribe) на определенные комнаты. ”Комнаты” полностью изолированы друг от друга, т.е. клиенты могут взаимодействовать только с “комнатами”, на которые они были подписаны ранее.  

Клиент получает историю каждой комнаты при успешном подключении. Каждый клиент должен иметь имя, имена должны быть уникальными в рамках “комнаты”.  
Список комнат сервера задается в файле config.json  
***
## Использование:
1 консоль - запуск сервер: В папке Server:  
go run server.go space.go rooms.pb.go  

2 консоль - использование клиента: в папке Client:  
go run client.go rooms.pb.go  

## Подписка:
Ввести номер комнаты для подписки, затем нажать Enter, для прекращения сеанса подписки клиента --end

## Публикация:
Будет выведено до 128 последних сообщений с указанием автора  
Затем можно ввести свои, закончить сеанс: --end  

Затем будет предложено повторить то же самое для следующего пользователя  
Отказаться: ввести --end  

## Тестирование:
В папке server:  
go test -v  

Тестирование включает проверки:
- запуска и остановки сервера  
- подписки клиента  
- публикации сообщений 