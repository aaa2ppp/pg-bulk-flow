#!/bin/sh

# Скрипт для объединения исходных файлов проекта
# Использование: ./merge_code.sh <директория1> <директория2> ...

for dir in "$@"; do
    # Ищем файлы с нужными расширениями, исключая тестовые миграции
    find "$dir" -type f \
        ! -path './migrations/test/*' \
        \( -name '*.go' -o -name '*.sql' -o -name '*.js' -o -name '*.sh' -o -name '*.md' \
        -o -name 'Dockerfile*' -o -name '*.y*ml' -o -name 'Makefile*' -o -name '*.example' \) \
    | while read -r f; do
        # Убираем ./ в начале пути
        f="${f#./}"
        
        # Выбираем разделитель в зависимости от типа файла
        case "$f" in
            *.go|*.js) sep="//" ;;
            *.sql)     sep="--" ;;
            *)         sep="#"  ;;
        esac
        
        # Выводим содержимое с заголовком
        echo "$sep == $f =="
        cat "$f"
        echo
    done
done
