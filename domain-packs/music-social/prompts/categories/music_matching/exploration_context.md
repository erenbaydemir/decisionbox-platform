## Music Matching App Context

This is a **music-taste-based social matching app**. Users connect their streaming service (Spotify, Apple Music, YouTube Music) and get matched with others who share similar music taste. Key aspects to explore:

- **Matching mechanics**: Users see match cards based on shared music preferences. Cards can show recently played tracks, playlists, or sponsored content. Users swipe right (like) or left (ignore) on profiles. Mutual right-swipes create a match. Super-likes and boosts increase visibility.
- **Match sources**: Matches come from different pools — instant matches (currently listening to the same music), recent matches (shared listening history), likes matches (someone already liked you), and artist room matches (met in a virtual listening room). Each source has different engagement dynamics.
- **Streaming integration**: The app connects to Spotify (premium and free tiers), Apple Music, and YouTube Music. The connected streaming service determines what music data is available for matching. Currently-playing detection enables real-time "instant matching" with other listeners.
- **Chat and connection**: After matching, users can chat via text, voice messages, and photo sharing. Some users share Instagram handles or phone numbers (tracked as tokens in messages). Quick reply templates and chat games are available to break the ice.
- **Social features**: Users can create social posts, participate in artist rooms (virtual listening sessions), answer community questions, and explore trending music/artists.
- **Premium/Elite model**: Freemium app with "Elite" subscription tier. Premium features include seeing who liked you (blurred matches), gender preference filters, unlimited swipes, profile boost, and priority matching. Multiple paywall versions are A/B tested.
- **Rewards and gamification**: Gift/reward system with missions (listen to songs, like artists/playlists/tracks) that unlock rewards. Incentivizes music exploration and platform engagement.
- **Anonymous matching**: An anonymous queue feature lets users get matched without revealing identity initially.

### Music Matching Example Queries

> **Important**: Adapt all column names, table names, and SQL functions to match the actual schema in {{SCHEMA_INFO}} and the connected warehouse's SQL dialect.

**Match Source Quality**:
```sql
-- Compare match-to-chat conversion by match source
-- Event parameters contain "from" field indicating the match source
SELECT
  JSON_EXTRACT_SCALAR(event_parameters, '$.from') as match_source,
  COUNT(DISTINCT user_id) as swipers,
  COUNT(*) as total_swipes
FROM `{{DATASET}}.events`
{{FILTER}}
  AND event_name = 'match_swipe_right_success'
GROUP BY match_source
ORDER BY total_swipes DESC
```

**Chat Engagement Depth**:
```sql
-- Analyze messaging patterns: how many messages do matched users exchange?
SELECT
  JSON_EXTRACT_SCALAR(event_parameters, '$.message_type') as msg_type,
  COUNT(*) as message_count,
  COUNT(DISTINCT user_id) as unique_senders
FROM `{{DATASET}}.events`
{{FILTER}}
  AND event_name = 'chat_message_sent'
GROUP BY msg_type
ORDER BY message_count DESC
```

**Music Exploration Patterns**:
```sql
-- How users discover music through the explore page
SELECT
  JSON_EXTRACT_SCALAR(event_parameters, '$.section') as explore_section,
  COUNT(*) as clicks,
  COUNT(DISTINCT user_id) as unique_users
FROM `{{DATASET}}.events`
{{FILTER}}
  AND event_name = 'explore_song_click'
GROUP BY explore_section
ORDER BY clicks DESC
```

**Streaming Service Engagement**:
```sql
-- Which streaming service drives more play-throughs?
SELECT
  event_name,
  COUNT(*) as plays,
  COUNT(DISTINCT user_id) as unique_players
FROM `{{DATASET}}.events`
{{FILTER}}
  AND event_name IN ('play_in_spotify', 'play_in_youtube_music', 'play_in_apple_music',
                     'play_in_spotify_sponsored', 'play_in_youtube_music_sponsored', 'play_in_apple_music_sponsored')
GROUP BY event_name
ORDER BY plays DESC
```

**Currently Playing and Instant Matching**:
```sql
-- How does currently-playing behavior drive instant matches?
SELECT
  JSON_EXTRACT_SCALAR(event_parameters, '$.instant_match_enabled') as instant_match_on,
  JSON_EXTRACT_SCALAR(event_parameters, '$.is_sponsored') as is_sponsored,
  COUNT(*) as events,
  COUNT(DISTINCT user_id) as unique_users
FROM `{{DATASET}}.events`
{{FILTER}}
  AND event_name = 'currently_playing_music_changed'
GROUP BY instant_match_on, is_sponsored
ORDER BY events DESC
```