-- 공통으로 사용될 tags 테이블
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    value VARCHAR(100) NOT NULL UNIQUE
);

-- INTEGER[] 사용
CREATE TABLE articles_with_int_array (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    tag_ids INTEGER[] NOT NULL DEFAULT '{}'
);

-- INTEGER[] 컬럼에 대한 GIN 인덱스
CREATE INDEX idx_articles_int_array_tags ON articles_with_int_array USING GIN (tag_ids);
