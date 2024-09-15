# Запуск этой реализации задания

1. Для обновления исполняемых файлов тестов на локалхосте сходи по ссылке https://github.com/Yandex-Practicum/go-autotests/releases/latest и скачай собранный бинарь
gophermarttest с тестами и statictest с анализатором.
Подробности: https://github.com/Yandex-Practicum/go-autotests
1. Для локального запуска тестов в корне проекта выполни
   1. `docker compose up -d`
   1. `docker compose exec cli bash`
   1. `go mod tidy; go mod vendor`
   1. `./vet.sh`
   1. `./test.sh`
1. Для отладки выполни
   1. внутри контейнера `dlv debug ./cmd/gophermart/main.go --listen :2345 --headless=true --api-version=2`
   1. снаружи контейнера подключиться к порту `40000` или к тому что указан в переменной `DELVE_EXTERNAL_PORT`

# go-musthave-diploma-tpl

Шаблон репозитория для индивидуального дипломного проекта курса «Go-разработчик»

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.
