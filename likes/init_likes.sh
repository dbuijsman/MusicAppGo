#!/usr/bin/env bash
echo Insert username:
read user
echo Insert password:
read -s pass
sudo mysql -u${user} -p${pass} <<EOF
CREATE DATABASE IF NOT EXISTS pref_likes;
USE pref_likes;
CREATE TABLE IF NOT EXISTS users (id INT NOT NULL PRIMARY KEY AUTO_INCREMENT, username VARCHAR(64) NOT NULL, UNIQUE(username));
CREATE TABLE IF NOT EXISTS artists (id INT NOT NULL PRIMARY KEY, name_artist VARCHAR(64) NOT NULL, prefix VARCHAR(7), UNIQUE(name_artist));
CREATE TABLE IF NOT EXISTS songs (id INT NOT NULL PRIMARY KEY, name_song VARCHAR(64) NOT NULL);
CREATE TABLE IF NOT EXISTS discography (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, artist_id INT NOT NULL, FOREIGN KEY (artist_id) REFERENCES artists (id) ON UPDATE CASCADE ON DELETE CASCADE, song_id INT NOT NULL, FOREIGN KEY (song_id) REFERENCES songs (id) ON UPDATE CASCADE ON DELETE CASCADE);
CREATE TABLE IF NOT EXISTS liked_songs (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, user_id INT NOT NULL, FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE, song_id INT NOT NULL, FOREIGN KEY (song_id) REFERENCES songs (id) ON UPDATE CASCADE ON DELETE CASCADE, UNIQUE(user_id, song_id));
CREATE TABLE IF NOT EXISTS disliked_songs (id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, user_id INT NOT NULL, FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE, song_id INT NOT NULL, FOREIGN KEY (song_id) REFERENCES songs (id) ON UPDATE CASCADE ON DELETE CASCADE, UNIQUE(user_id, song_id));
CREATE USER IF NOT EXISTS likesMusicApp IDENTIFIED BY 'likelikes';
GRANT SELECT, INSERT ON pref_likes.users TO likesMusicApp;
GRANT SELECT, INSERT ON pref_likes.artists TO likesMusicApp;
GRANT SELECT, INSERT ON pref_likes.songs TO likesMusicApp;
GRANT SELECT, INSERT ON pref_likes.discography TO likesMusicApp;
GRANT SELECT, INSERT, DELETE ON pref_likes.liked_songs TO likesMusicApp;
GRANT SELECT, INSERT, DELETE ON pref_likes.disliked_songs TO likesMusicApp;
EOF
