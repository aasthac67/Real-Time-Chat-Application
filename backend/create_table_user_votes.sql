CREATE TABLE user_votes (
    user_id VARCHAR(255),
    message_id VARCHAR(255),
    vote_type VARCHAR(20), -- 'upvote' or 'downvote'
    PRIMARY KEY (user_id, message_id)
);