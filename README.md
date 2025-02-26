# gRPC File Service

## Описание проекта

Проект `tages` представляет собой сервис на Go, который использует gRPC для загрузки, скачивания и просмотра списка бинарных файлов (изображений). Сервис ограничивает количество одновременных подключений:
- на загрузку/скачивание файлов — 10 конкурентных запросов;
- на просмотр списка файлов — 100 конкурентных запросов.

## Структура проекта
tages/
├── cmd/
│ ├── server/
│ │ └── main.go
│ └── client/
│ └── main.go
├── pkg/
│ └── proto/
│ ├── file_service.pb.go
│ └── file_service_grpc.pb.go
│ ├── service/
│ │ └── service.go
│ └── storage/
│ └── storage.go
├── proto/
│ └── file_service.proto
├── go.mod
└── go.sum


## Требования

- Go 1.20 или новее
- `protoc` компилятор
- Плагины `protoc-gen-go` и `protoc-gen-go-grpc`

## Установка

1. **Скачайте и установите `protoc`**:
    - Перейдите на [страницу релизов `protoc`](https://github.com/protocolbuffers/protobuf/releases).
    - Скачайте последнюю версию для Windows (например, `protoc-3.21.12-win64.zip`).
    - Распакуйте архив в удобную для вас директорию, например, `C:\protoc`.
    - Добавьте путь `C:\protoc\bin` в переменную окружения `PATH`.

2. **Установите плагины `protoc-gen-go` и `protoc-gen-go-grpc`**:
    ```sh
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```

3. **Сгенерируйте код Go из `.proto` файла**:
    ```sh
    protoc --go_out=./pkg/proto --go-grpc_out=./pkg/proto proto/file_service.proto
    ```

4. **Инициализируйте модуль Go**:
    ```sh
    go mod init local/tages
    ```

5. **Установите зависимости**:
    ```sh
    go get google.golang.org/grpc
    go get google.golang.org/protobuf/cmd/protoc-gen-go
    go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
    ```

## Запуск сервера

1. Перейдите в директорию проекта:
    ```sh
    cd C:\Users\RSapunov\GolandProjects\tages
    ```

2. Запустите сервер:
    ```sh
    go run cmd/server/main.go
    ```

## Запуск клиента

### Загрузка файла

```sh
go run cmd/client/main.go -action upload -filename qwerty.jpg
