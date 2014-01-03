CREATE TABLE github_builds(
  access_token VARCHAR(100) primary key,
  owner VARCHAR(100) NOT NULL,
  repository VARCHAR(100) NOT NULL
);
