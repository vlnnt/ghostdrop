# GhostDrop

**Язык:** [English](README.md) | Русский

Простой Discord-бот для анонимной отправки изображений, GIF, видео и аудиофайлов.

## Возможности

- Анонимная отправка медиа через `/anon`.
- Поддержка изображений, GIF, видео и аудио.
- Выбор любого текстового канала.
- Простая настройка через веб-интерфейс.

## Установка

1. Скачайте последнюю версию в разделе [Releases](../../releases).
2. Создайте бота в [Discord Developer Portal](https://discord.com/developers/applications).
3. Пригласите бота на сервер с разрешениями `bot` и `applications.commands`.
4. Запустите `ghostdrop.exe` на Windows или `./ghostdrop` на Linux.
5. Откройте `http://127.0.0.1:8080` и введите токен бота.
6. Нажмите **Save & Start GhostDrop**.

## Docker

**Сборка:**

```bash
sudo docker build -f docker/Dockerfile -t ghostdrop .
```

**Запуск:**

```bash
sudo docker run -d \
  --name ghostdrop \
  --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -v ghostdrop-data:/data \
  ghostdrop
```

## Использование

```text
/anon media:<файл> channel:<канал>
```

**Пример:**

```text
/anon media:funny.gif channel:#general
```

Бот отправит файл в выбранный канал, не раскрывая публично личность отправителя.

**Поддерживаемые форматы:** PNG, JPG, GIF, WebP, MP4, WebM, MOV, MP3, WAV, OGG, FLAC и M4A.

**Максимальный размер файла:** 20 МиБ.

## Настройка

Настройки автоматически сохраняются в `config.yaml` рядом с исполняемым файлом или в `/data/config.yaml` при использовании Docker.

Оставьте Guild ID пустым, чтобы команды работали на всех серверах, где установлен бот.

Не передавайте другим людям токен бота и файл конфигурации.