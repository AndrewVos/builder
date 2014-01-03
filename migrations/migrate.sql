CREATE TABLE github_builds(
  access_token VARCHAR(100) NOT NULL,
  owner VARCHAR(100) NOT NULL,
  repository VARCHAR(100) NOT NULL
);

CREATE TABLE builds(
  build_id   VARCHAR(100) PRIMARY KEY,
  url        VARCHAR(200),
  owner      VARCHAR(100),
  repository VARCHAR(100),
  ref        VARCHAR(100),
  sha        VARCHAR(50),
  complete   BOOLEAN,
  success    BOOLEAN,
  result     VARCHAR(30),
  github_url VARCHAR(200)
);

CREATE TABLE commits(
  build_id VARCHAR(100),
  sha      VARCHAR(50),
  message  TEXT,
  url      VARCHAR(200)
);
