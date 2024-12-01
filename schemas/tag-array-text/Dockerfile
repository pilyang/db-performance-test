#PostgreSQL 16 버전을 베이스 이미지로 사용
FROM postgres:16

COPY ./init/* /docker-entrypoint-initdb.d/

#계정정보 설정
ENV POSTGRES_USER=test
ENV POSTGRES_PASSWORD=mypassword
ENV POSTGRES_DB=test_db

EXPOSE 5432
