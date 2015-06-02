## Feed прокси для поинтопараши

Сообщения об ошибках и пожелания можно оставлять [тут](https://github.com/etw/pointfeed/issues)

### Работающие фиды:

* /feed/all - Все подряд
* /feed/tags?tag=\<tag1\>&tag=\<tag2\>&... - Посты с хотя бы одним из указанных тегов

### Универсальные GET параметры
* before=\<uid\> - начать фид с указанного поста
* minposts=\<num\> - сформировать фид из не менее, чем указанное количество постов (по-дефолту 20)
* nouser=\<name\> - не включать в фид посты пользователя (можно применять несколько раз)
* notag=\<tag\> - не включать в фид посты с тегом (можно применять несколько раз)

### Фичи

* Поддержка работы с поинтом через SOCKS-прокси
* Поддержка Markdown разметки
* Вставка прикрепленных к постам картинок
* Управление попаданием своих постов в фиды (через BL/WL пользователя @feed)
* Возможность получать более 20 последних постов в фиде
* Исключение постов по фильтрам (пользователи / теги)

### ToDo

* Автозамена ссылок на безопасные (http -> https)
* Показ изображений вместо прямых ссылок на них
* Показ изображений из постов на gelbooru

#### Слава BnW!
