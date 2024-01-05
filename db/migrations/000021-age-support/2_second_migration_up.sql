-- NOTE: comments preceding each CREATE section are to be utilized to create additional edges once referenced nodes are created
-- add genre information for ambient music
SET search_path = ag_catalog, "$user", public;

-- influenced_by: Krautrock, Musique Concrete, Minimalism, Free Jazz, Drone, Field Recordings, Industrial
SELECT * FROM cypher('music_genres', $$
  CREATE (ambient:Genre {name: "Ambient", type: "base_genre"}),
  (ambient_description:Genre_description {language: "en",
    description: "Ambient is a genre of music that emphasizes tone and atmosphere over traditional musical structure or rhythm.
    A form of instrumental music, it may lack neticable beats, meter, key and structured melody.
    It uses textural layers of sound which can reward both passive and active listening and encourage a sense of calm or contemplation.
    The genre is said to evoke an \"atmospheric\", \"visual\", or \"unobtrusive\" quality.
    Nature soundscapes may be included, and the sounds of acoustic instruments such as the piano, strings and flute may be emulated through a synthesizer."}),
  (ambient)-[:HAS_DESCRIPTION]->(ambient_description)
  SET ambient.characteristics = ['atmospheric', 'visual', 'unobtrusive', 'calm', 'contemplative', 'minimalistic', 'background']
  $$) as (result agtype);

-- influnced_by: Harsh Noise, Drone, Field Recordings, Industrial, Musique Concrete
SELECT * FROM cypher('music_genres', $$
  CREATE (ambient_noise_wall:Genre {name: "Ambient Noise Wall", type: "sub_genre"}),
  (ambient_noise_wall_description:Genre_description {language: "en",
    description: "A combination of Harsh Noise Wall and ambient music."}),
  (ambient_noise_wall)-[:HAS_DESCRIPTION]->(ambient_noise_wall_description),
  (ambient)-[:HAS_SUBGENRE]->(ambient_noise_wall)
  SET ambient_noise_wall.characteristics = ['monolithic', 'distorted', 'minimalistic', 'harsh']
$$) as (result agtype);

-- fusion_genres: Ambient Black Metal, Martial Industrial, Funeral Doom
-- influenced_by: Industrial, Ambient, Drone, Noise
SELECT * FROM cypher('music_genres', $$
  CREATE (dark_ambient:Genre {name: "Dark Ambient", type: "sub_genre"}),
  (dark_ambient_description:Genre_description {language: "en",
    description: "Dark ambient (especially in the 1980s referred to as ambient industrial) is a genre of post-industrial music that features an ominous, 
    dark droning and often gloomy, monumental or catacombal atmosphere, partially with discordant overtones.
    It shows similarities towards ambient music, a genre that has been cited as a main influence by many dark ambient artists, both conceptually and compositionally.
    Although mostly electronically generated, dark ambient also includes the sampling of hand-played instruments and semi-acoustic recording procedures, and is strongly related to ritual industrial or ritual ambient."}),
  (dark_ambient)-[:HAS_DESCRIPTION]->(dark_ambient_description),
  (ambient)-[:HAS_SUBGENRE]->(dark_ambient)
  SET dark_ambient.characteristics = ['ominous', 'dark', 'gloomy', 'monumental', 'catacombal', 'discordant', 'ritualistic', 'industrial']
$$) as (result agtype);

-- influence_by: Black Metal, Dark Ambient, Drone, Noise, Field Recordings
SELECT * FROM cypher('music_genres', $$
  CREATE (black_ambient:Genre {name: "Black Ambient", type: "sub_genre"}),
  (black_ambient_description:Genre_description {language: "en",
    description: "Black ambient is a subgenre of dark ambient that features additional influences from black metal and/or darkwave.
    The genre emerged in the early 1990s with the works of artists such as Brighter Death Now, Robert Fripp, Lustmord, Nocturnal Emissions, Zoviet France, and Z'EV.
    A similar genre called blackened ambient also exists, which is a fusion of black metal and dark ambient."}),
  (black_ambient)-[:HAS_DESCRIPTION]->(black_ambient_description),
  (dark_ambient)-[:HAS_SUBGENRE]->(black_ambient)
  SET black_ambient.chracteristics = ['ritualistic', 'gloomy', 'dark', 'dissonant', 'distorted', 'field recordings', 'melancholic', 'organic']
$$) as (result agtype);

-- influnced_by: World Music, Dark Ambient, Religious Music, Drone, Tribal Ambient
SELECT * FROM cypher('music_genres', $$
  CREATE (ritual_ambient:Genre {name: "Ritual Ambient", type: "sub_genre"}),
  (ritual_ambient_description:Genre_description {language: "en",
    description: "Closely related to dark ambient, it involves ritualistic elements and/or influences from world music."}),
    (ritual_ambient)-[:HAS_DESCRIPTION]->(ritual_ambient_description),
    (dart_ambient)-[:HAS_SUBGENRE]->(ritual_ambient)
    SET ritual_ambient.characteristics = ['ritualistic', 'world music', 'dark', 'deep', 'mystical', 'disturbing', 'tribal', 'meditative', 'hypnotic']
$$) as (result agtype);

-- influnced_by: Space Rock, Dark Ambient, Drone, Psychedelic Rock, Progressive Rock, Heavy Psych
SELECT * FROM cypher('music_genres', $$
  CREATE (space_ambient:Genre {name: "Space Ambient", type: "sub_genre"}),
  (space_ambient_description:Genre_description {language: "en",
    description: "Space ambient is a subgenre of dark ambient that evokes feelings of outer space.
    The genre emerged in the 1990s with the works of artists such as Steve Roach, Robert Rich, and Michael Stearns.
    It has its roots in 1970s ambient music and Brian Eno's works with ambient music, but is more specifically influenced by the space music subgenre.
    The genre is also related to space rock, a fusion of progressive rock and psychedelic rock."}),
  (space_ambient)-[:HAS_DESCRIPTION]->(space_ambient_description),
  (ambient)-[:HAS_SUBGENRE]->(space_ambient)
  SET space_ambient.characteristics = ['space', 'cosmic', 'atmospheric', 'ethereal', 'dreamy', 'floating', 'hypnotic', 'meditative', 'minimalistic', 'organic', 'psychedelic', 'spacy', 'trippy', 'relaxing']
$$) as (result agtype);

-- influenced_by: World Music, Dark Ambient, Ritual Ambient
SELECT * FROM cypher('music_genres',
  CREATE (tribal_ambient:Genre {name: "Tribal Ambient", type: "sub_genre"}),
  (tribal_ambient_description:Genre_description {language: "en",
    description: "Tribal ambient is a subgenre of dark ambient which combines dark ambient with tribal influences,
    using traditional instruments and musical structures from cultures around the world.
    The genre was pioneered by acts such as O Yuki Conjugate, Zoviet France, and Nocturnal Emissions."}),
  (tribal_ambient)-[:HAS_DESCRIPTION]->(tribal_ambient_description),
  (ambient)-[:HAS_SUBGENRE]->(tribal_ambient)
  SET space_ambient.characteristics = ['tribal', 'traditional', 'world music', 'atmospheric', 'deep', 'mystical']
$$) as (result agtype);
