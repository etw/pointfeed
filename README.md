## Feed прокси для поинтопараши

[Багтрекер](https://github.com/etw/pointfeed/issues)

[Исходники](https://github.com/etw/pointfeed/)

### Работающие URL-ы:

* /feed/all — все подряд;
* /feed/tags?tag=\<tag1\>&tag=\<tag2\>&... — посты с хотя бы одним из указанных тегов.

* /stats/ — статистика

### Универсальные GET параметры
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

* поддержка работы с поинтом через SOCKS-прокси (arts не спалит ОйПи!);
* поддержка Markdown разметки;
* вставка прикрепленных к постам картинок;
* управление попаданием своих постов в фиды (через BL/WL пользователя [@feed](https://feed.point.im));
* возможность получать более 20 последних постов в фиде;
* исключение постов по фильтрам (пользователи / теги);
* автозамена ссылок на безопасные (http -> https);
* показ изображений вместо прямых ссылок на них;
* показ изображений из постов на gelbooru;
* кеширование отрендереных постов.

#### Слава BnW!
