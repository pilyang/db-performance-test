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

-- INTEGER[] 방식의 데이터 생성
WITH article_numbers AS (
    SELECT generate_series AS num FROM generate_series(1, 100000)
)
INSERT INTO articles_with_int_array (title, tag_ids)
SELECT 
    'Article ' || a.num || ' ' || md5(random()::text),
    (
        SELECT array_agg(id)
        FROM (
            SELECT id
            FROM tags
            -- 각 게시글마다 다른 시드값 사용
            ORDER BY random() * a.num  
            LIMIT 3 + (random() < 0.5)::int
        ) t
    )
FROM article_numbers a;

-- INTEGER[] 방식 데이터 확인
SELECT 
    id, 
    title, 
    tag_ids,
    (
        SELECT array_agg(value)
        FROM tags
        WHERE id = ANY(tag_ids)
    ) as tag_values
FROM articles_with_int_array
ORDER BY id
LIMIT 5;
