-- articles 테이블 생성
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL
);

-- tags 테이블 생성
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    value VARCHAR(100) NOT NULL UNIQUE
);

-- article_tags 중간 테이블 생성
CREATE TABLE article_tags (
    article_id INTEGER REFERENCES articles(id),
    tag_id INTEGER REFERENCES tags(id),
    PRIMARY KEY (article_id, tag_id)
);

-- 인덱스 생성
CREATE INDEX idx_articles_title ON articles(title);
CREATE INDEX idx_tags_value ON tags(value);
CREATE INDEX idx_article_tags_tag_article ON article_tags(tag_id, article_id);
