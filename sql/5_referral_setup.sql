CREATE TABLE IF NOT EXISTS referral
(
    id                  INTEGER NOT NULL,
    vid                 INTEGER NOT NULL,
    where_to_refer      TEXT,
    additional_comments TEXT,

    PRIMARY KEY (id, vid),
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE
);
