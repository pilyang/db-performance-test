-- 태그 데이터 추가 (100개의 개발 관련 태그)
INSERT INTO tags (value) VALUES
    ('javascript'), ('python'), ('java'), ('cpp'), ('golang'),
    ('react'), ('vue'), ('angular'), ('svelte'), ('jquery'),
    ('nodejs'), ('deno'), ('typescript'), ('kotlin'), ('swift'),
    ('spring'), ('django'), ('flask'), ('fastapi'), ('express'),
    ('postgresql'), ('mysql'), ('mongodb'), ('redis'), ('elasticsearch'),
    ('docker'), ('kubernetes'), ('aws'), ('gcp'), ('azure'),
    ('ci-cd'), ('jenkins'), ('github-actions'), ('gitlab'), ('bitbucket'),
    ('testing'), ('jest'), ('pytest'), ('junit'), ('selenium'),
    ('react-native'), ('flutter'), ('android'), ('ios'), ('mobile'),
    ('html'), ('css'), ('sass'), ('tailwind'), ('bootstrap'),
    ('webpack'), ('vite'), ('babel'), ('eslint'), ('prettier'),
    ('graphql'), ('rest-api'), ('grpc'), ('websocket'), ('mqtt'),
    ('linux'), ('unix'), ('windows'), ('macos'), ('ubuntu'),
    ('git'), ('svn'), ('mercurial'), ('clean-code'), ('design-patterns'),
    ('agile'), ('scrum'), ('kanban'), ('devops'), ('sre'),
    ('security'), ('oauth'), ('jwt'), ('https'), ('ssl'),
    ('machine-learning'), ('ai'), ('deep-learning'), ('data-science'), ('numpy'),
    ('pandas'), ('tensorflow'), ('pytorch'), ('keras'), ('scikit-learn'),
    ('microservices'), ('serverless'), ('firebase'), ('heroku'), ('netlify'),
    ('debugging'), ('profiling'), ('monitoring'), ('logging'), ('analytics');

-- articles 데이터 추가
INSERT INTO articles (title)
SELECT 
    'Article ' || generate_series || ' ' || md5(random()::text)
FROM generate_series(1, 100000);

-- article_tags 관계 데이터 생성
WITH article_numbers AS (
    SELECT 
        a.id as article_id,
        row_number() OVER (ORDER BY a.id) as num
    FROM articles a
),
article_tags_random AS (
    SELECT 
        an.article_id,
        (SELECT array_agg(id)
         FROM (
             SELECT id
             FROM tags
             ORDER BY random() * an.num
             LIMIT 3 + (random() < 0.5)::int
         ) t
        ) as tag_ids
    FROM article_numbers an
)
INSERT INTO article_tags (article_id, tag_id)
SELECT article_id, unnest(tag_ids)
FROM article_tags_random;

-- 결과 확인
SELECT 
    article_id,
    COUNT(*) as tag_count
FROM article_tags
GROUP BY article_id
ORDER BY article_id
LIMIT 10;
