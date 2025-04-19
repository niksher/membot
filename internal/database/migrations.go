package database

type Migration struct {
	Name     string
	Commands []string
}

var migrations = []Migration{
	{
		Name: "01_create_tables",
		Commands: []string{
			`CREATE TABLE IF NOT EXISTS videos (
				id INT AUTO_INCREMENT PRIMARY KEY,
				file_id VARCHAR(255) NOT NULL UNIQUE,
				caption TEXT,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB`,

			`CREATE TABLE IF NOT EXISTS tags (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(50) NOT NULL UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB`,

			`CREATE TABLE IF NOT EXISTS video_tags (
				video_id INT NOT NULL,
				tag_id INT NOT NULL,
				PRIMARY KEY (video_id, tag_id),
				CONSTRAINT fk_video_tags_video 
					FOREIGN KEY (video_id) REFERENCES videos(id) 
					ON DELETE CASCADE,
				CONSTRAINT fk_video_tags_tag
					FOREIGN KEY (tag_id) REFERENCES tags(id)
					ON DELETE CASCADE
			) ENGINE=InnoDB`,

			`CREATE TABLE IF NOT EXISTS sent_videos (
				chat_id BIGINT NOT NULL,
				video_id INT NOT NULL,
				sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (chat_id, video_id),
				CONSTRAINT fk_sent_videos_video
					FOREIGN KEY (video_id) REFERENCES videos(id)
					ON DELETE CASCADE
			) ENGINE=InnoDB`,
		},
	},
}
