INSERT INTO tags (label) VALUES ("safe"), ("unsafe"), ("old"), ("young"), ("maybe");
INSERT INTO images (dhash, path, size) VALUES (12345, "temp/png.jpeg", 169);
INSERT INTO images (dhash, path, size) VALUES (67890, "temp/webp.jpeg", 420);
INSERT INTO images (dhash, path, size) VALUES (54321, "temp/jpeg.jpeg", 269);
INSERT INTO image_tags (dhash, label) VALUES (67890, "safe"), (67890, "old"), (54321, "safe"), (54321, "old");