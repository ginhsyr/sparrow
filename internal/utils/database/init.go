package database

import (
	"Sparrow/internal/model"
	"fmt"
	"gorm.io/gorm"
	"log"
)

func DBInit(db *gorm.DB) {
	models := []interface{}{
		&model.User{},
		&model.Post{},
		&model.PostContent{},
		&model.PostEdit{},
		// &model.PostLike{},
	}

	for _, m := range models {
		if !db.Migrator().HasTable(m) {
			if err := db.Migrator().CreateTable(m); err != nil {
				log.Printf("Create table err: %v", err)
			}
		}
	}

	if !db.Migrator().HasColumn(&model.Post{}, "like_count") {
		if err := db.Migrator().AddColumn(&model.Post{}, "LikeCount"); err != nil {
			log.Fatal(err)
		}
	}

	if !db.Migrator().HasTable(&model.PostLike{}) {
		if err := db.Exec(`CREATE TABLE post_likes (
	    		post_id BIGINT NOT NULL,
	    		user_id BIGINT NOT NULL,
	    		liked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    		PRIMARY KEY (post_id, user_id),
				CONSTRAINT fk_post_likes_post FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
				CONSTRAINT fk_post_likes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			) PARTITION BY HASH (post_id)`).Error; err != nil {
			log.Fatal(err)
		}
	}

	if err := EnsurePostLikeConstraints(db); err != nil {
		log.Fatal(err)
	}

	if !db.Migrator().HasConstraint(&model.PostEdit{}, "Post") {
		err := db.Migrator().CreateConstraint(&model.PostEdit{}, "Post")
		if err != nil {
			log.Fatal(err)
		}
	}

	if !db.Migrator().HasConstraint(&model.PostEdit{}, "Editor") {
		err := db.Migrator().CreateConstraint(&model.PostEdit{}, "Editor")
		if err != nil {
			log.Fatal(err)
		}
	}

	err := EnsurePostLikePartitions(db, 64)
	if err != nil {
		log.Fatal(err)
	}
}

func EnsurePostLikeConstraints(db *gorm.DB) error {
	constraints := []string{
		`
		DO $$
		BEGIN
		    IF NOT EXISTS (
		        SELECT 1
		        FROM pg_constraint
		        WHERE conname = 'fk_post_likes_post'
		          AND conrelid = 'post_likes'::regclass
			    ) THEN
			        ALTER TABLE post_likes
			            ADD CONSTRAINT fk_post_likes_post
			            FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE;
			    END IF;
			END $$;
			`,
		`
		DO $$
		BEGIN
		    IF NOT EXISTS (
		        SELECT 1
		        FROM pg_constraint
		        WHERE conname = 'fk_post_likes_user'
		          AND conrelid = 'post_likes'::regclass
			    ) THEN
			        ALTER TABLE post_likes
			            ADD CONSTRAINT fk_post_likes_user
			            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			    END IF;
			END $$;
			`,
	}

	for _, sql := range constraints {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func EnsurePostLikePartitions(db *gorm.DB, numPartitions int) error {
	existing, err := getExistingPostLikePartitions(db)
	if err != nil {
		return err
	}

	for i := 0; i < numPartitions; i++ {
		if _, ok := existing[i]; ok {
			continue
		}
		tableName := fmt.Sprintf("post_likes_p%d", i)
		createSQL := fmt.Sprintf(`
	            CREATE TABLE IF NOT EXISTS %s PARTITION OF post_likes
	            FOR VALUES WITH (modulus %d, remainder %d);
	        `, tableName, numPartitions, i)
		if err := db.Exec(createSQL).Error; err != nil {
			return err
		}
	}
	return nil
}

func getExistingPostLikePartitions(db *gorm.DB) (map[int]struct{}, error) {
	var indices []int
	query := `
		SELECT CAST(REPLACE(child.relname, 'post_likes_p', '') AS INTEGER) AS partition_index
		FROM pg_inherits
		JOIN pg_class parent ON pg_inherits.inhparent = parent.oid
		JOIN pg_class child ON pg_inherits.inhrelid = child.oid
		WHERE parent.relname = 'post_likes'
		  AND child.relname ~ '^post_likes_p[0-9]+$'
	`
	if err := db.Raw(query).Scan(&indices).Error; err != nil {
		return nil, err
	}

	existing := make(map[int]struct{}, len(indices))
	for _, idx := range indices {
		existing[idx] = struct{}{}
	}
	return existing, nil
}
