# Локальные Protobuf зависимости

Эта директория содержит локальные копии protobuf зависимостей, необходимых для генерации protobuf файлов.

## Содержимое

- `googleapis/` - Google API protobuf определения
- `grpc-gateway/` - gRPC Gateway protobuf определения

## Обновление зависимостей

Для обновления зависимостей выполните:

```powershell
# Обновить googleapis
cd proto-deps/googleapis
git pull origin main
cd ../..

# Обновить grpc-gateway
cd proto-deps/grpc-gateway
git pull origin main
cd ../..
```

Или удалите директории и пересоздайте их:

```powershell
Remove-Item -Recurse -Force proto-deps/googleapis
Remove-Item -Recurse -Force proto-deps/grpc-gateway
git clone --depth 1 https://github.com/googleapis/googleapis.git proto-deps/googleapis
git clone --depth 1 https://github.com/grpc-ecosystem/grpc-gateway.git proto-deps/grpc-gateway
```

## Примечание

Эти зависимости используются вместо удаленных модулей buf.build для генерации protobuf файлов, когда доступ к buf.build недоступен.

