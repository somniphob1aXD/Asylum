# 🎵 Asylum

Терминальный плеер для SoundCloud с красивым ASCII-артом, системой избранного и автоматическим переключением треков.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-blue?style=for-the-badge)

## ✨ Возможности

- 🔍 Поиск треков и плейлистов на SoundCloud
- 👻 Красивый ASCII-арт Gengar вместо обложек
- ⭐ Система избранных треков (сохраняется между сессиями)
- ⏭️ Автоматическое переключение треков
- ️ Управление громкостью и перемоткой
- 🔄 Поддержка SOCKS5/HTTP прокси (для регионов с блокировками)
- 💾 Локальное кэширование треков

##  Установка

### Зависимости

Для работы требуются `yt-dlp` и `ffmpeg`:

**Arch Linux / CachyOS:**

    sudo pacman -S yt-dlp ffmpeg

**Ubuntu / Debian:**

    sudo apt install yt-dlp ffmpeg

**macOS:**

    brew install yt-dlp ffmpeg

**Windows (через Chocolatey):**

    choco install yt-dlp ffmpeg

### Компиляция из исходников

    # Клонируйте репозиторий
    git clone https://github.com/ВАШ_НИК/asylum.git
    cd asylum

    # Установите Go зависимости
    go mod download

    # Скомпилируйте
    go build -o asylum main.go

    # Установите в систему (опционально)
    sudo install -m 755 asylum /usr/local/bin/asylum

## 🚀 Использование

### Базовый запуск

    # Если SoundCloud доступен в вашем регионе
    asylum

    # Если SoundCloud заблокирован (РФ, СНГ)
    export SC_PROXY="socks5://127.0.0.1:1080"
    asylum

### Настройка прокси

Добавьте в `~/.bashrc` или `~/.zshrc`:

    export SC_PROXY="socks5://127.0.0.1:1080"

## ⌨️ Управление

| Клавиша | Действие |
|---------|----------|
| `/` | Поиск треков |
| `F` | Режим избранного |
| `f` | Добавить/убрать трек из избранного |
| `j` / `k` | Навигация вверх / вниз |
| `Enter` | Воспроизвести выбранный трек |
| `Space` / `p` | Пауза / продолжить |
| `h` / `l` | Перемотка назад / вперёд (±5 сек) |
| `n` / `N` | Следующий / предыдущий трек |
| `+` / `-` | Увеличить / уменьшить громкость |
| `Esc` | Выйти из режима избранного / закрыть поиск |
| `q` / `Ctrl+C` | Выход из программы |

##  Структура проекта

    asylum/
    ├── main.go              # Основной код приложения
    ├── gengar.txt           # ASCII-арт Gengar (встраивается в бинарник)
    ├── go.mod               # Go модуль
    ├── go.sum               # Зависимости Go
    ├── README.md            # Документация
    └── .gitignore           # Игнорируемые файлы

## 💾 Хранение данных

Все данные сохраняются в папке `.sc-tui-cache/` в директории запуска:

- `favorites.json` — список избранных треков
- `track-*.mp3` — временно скачанные треки (удаляются после воспроизведения)

## 🎨 ASCII-арт

Хотите изменить на другой арт? Просто отредактируйте файл `gengar.txt` и пересоберите:

    nano gengar.txt  # Вставьте новый ASCII-арт
    go build -o asylum main.go
    sudo install -m 755 asylum /usr/local/bin/asylum

## ️ Разработка

    # Запуск в режиме разработки
    go run main.go

    # Сборка с отладочной информацией
    go build -gcflags="all=-N -l" -o asylum main.go

    # Запуск тестов (если будут)
    go test ./...

## 🤝 Вклад

Pull Request'ы приветствуются! Если вы нашли баг или хотите предложить новую функцию — создайте Issue.

## 📄 Лицензия

MIT License — см. файл [LICENSE](LICENSE) для деталей.

## 🙏 Благодарности

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI фреймворк
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) — загрузка аудио
- [beep](https://github.com/faiface/beep) — аудиоплеер для Go
- [go-runewidth](https://github.com/mattn/go-runewidth) — корректное отображение braille-символов

---

**Made ancientreligionbtw #lowtier**
