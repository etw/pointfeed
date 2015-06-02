## Feed прокси для поинтопараши

Сообщения об ошибках и пожелания можно оставлять [тут](https://github.com/etw/pointfeed/issues).

### Работающие фиды:

* /feed/all — Все подряд;
* /feed/tags?tag=\<tag1\>&tag=\<tag2\>&... — посты с хотя бы одним из указанных тегов.

### Универсальные GET параметры
* before=\<uid\> — начать фид с указанного поста;
* minposts=\<num\> — сформировать фид из не менее, чем указанное количество постов (по-дефолту 20);
* nouser=\<name\> — не включать в фид посты пользователя (можно применять несколько раз);
* notag=\<tag\> — не включать в фид посты с тегом (можно применять несколько раз).

#### Примеры

https://pointfeed-etw.rhcloud.com/feed/all?nouser=radjah&nouser=Nico-izo
// Любитель китайских порномультиков - не человек

https://pointfeed-etw.rhcloud.com/feed/tags?tag=manga&notag=anime
// Экранизации сосут

https://pointfeed-etw.rhcloud.com/feed/tags?tag=anime_art&nouser=animal-love
// Без собак, плиз

### Фичи

* поддержка работы с поинтом через SOCKS-прокси;
* поддержка Markdown разметки;
* вставка прикрепленных к постам картинок;
* управление попаданием своих постов в фиды (через BL/WL пользователя [@feed](https://feed.point.im));
* возможность получать более 20 последних постов в фиде;
* исключение постов по фильтрам (пользователи / теги).

### ToDo

* автозамена ссылок на безопасные (http -> https);
* показ изображений вместо прямых ссылок на них;
* показ изображений из постов на gelbooru.

#### Слава BnW!
