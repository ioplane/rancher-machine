# Восстановление незавершённого Create

Форк добавляет необязательный `drivers.CreateRecoveryDriver` и RPC
`PrepareCreateRecovery`. Обязательный `drivers.Driver`, версия RPC и версия
`config.json` сохранены. Для старого RPC-сервера только точный ответ
`rpc.ServerError("rpc: can't find method <service>.PrepareCreateRecovery")`
означает отсутствие capability; транспортные ошибки и ошибки драйвера
возвращаются вызывающей стороне.

Перед отказом существующему хосту команда передаёт текущие driver flags,
наличие каталога, путь custom script и hostname override. Используется новый
Host текущего запуска; сохранённые credentials не загружаются через `Load`.
Capability вызывается только для custom script через SSH без userdata-флага.
Обычная установка Docker, Engine/Swarm/Auth provisioning и userdata остаются
за пределами recovery. Для неподдерживаемого драйвера, завершённого,
удалённого или legacy-хоста сохраняется `ErrHostAlreadyExists` до чтения
custom script и создания временного userdata-файла.

Драйвер самостоятельно доказывает принадлежность pending-журнала, UUID,
SSH-пары и текущего intent. Ответ `true` означает, что текущие flags уже
проверены и установлены; повторный `SetConfigFromFlags` пропускается.
Capability не разрешает повтор VM POST. `PreCreateCheck`, Save перед Create,
Create и Save после Create выполняются существующим путём.

Совместимый Nutanix-драйвер связывает checkpoint с digest режима
`ssh-custom-script/v1`, содержимого script и эффективного hostname.
Опциональный persisted digest требуется последующим обычным GetState/Remove;
raw script и credentials в proof не сохраняются. Старый consumer не создаёт
такой checkpoint. Опубликованный fork build обоих компонентов требуется
совместно; регистрация в Rancher, bootstrap гостя и production canary являются
отдельными этапами приёмки.

## Проверка

`make dev-build lint-build check` выполняет проверки в Podman с закреплённым
Go 1.27.2 и golangci-lint v2.14.0. Race-тесты ограничены критическими create/RPC/logger-тестами в commands,
libmachine/drivers/rpc и libmachine/crashreport;
vet/build/module integrity/vulnerability scan охватывают весь модуль. Линтер
проверяет новые замечания относительно исходного commit
`407bfdc12e6fdfd19c9d76d565d552ea5b39f363`, formatter — новые Go-файлы;
унаследованный код провайдеров массово не переформатируется.

Расширенная историческая suite отдельно выявила baseline-отказы:
`TestIP` ожидает `IP address is not set`, реализация возвращает
`IPv4 address is not set`; `TestGetScpCmdWithDelta` требует `rsync`, которого
нет в минимальном dev-образе. Эти проверки не входят в критический regex
`make check-test`; общий исторический набор не заявлен как прошедший.
