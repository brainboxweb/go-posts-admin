DROP TABLE posts;
CREATE TABLE `posts` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `slug` VARCHAR(255) NULL,
    `title` VARCHAR(255) NULL,
    `description` VARCHAR(400) NULL,
    `published` DATETIME NULL,
    `body` TEXT,
    `transcript` TEXT NULL,
    `topresult` TEXT NULL,
    `click_to_tweet` VARCHAR(20)
 );


DROP TABLE `posts_keywords_xref`;
CREATE TABLE `posts_keywords_xref` (
  `post_id` INT,
  `keyword_id` VARCHAR(100),
  PRIMARY KEY (post_id, keyword_id)
);



DROP TABLE `youtube`;
CREATE TABLE `youtube` (
  `id` VARCHAR(255) PRIMARY KEY,
  `post_id` INT NOT NULL,
  `body` TEXT NULL
);


DROP TABLE `youtube_music_xref`;
CREATE TABLE `youtube_music_xref` (
    `youtube_id` INT,
    `music_id` VARCHAR(255)
);
