SELECT create_graph('music_genres');
SET search_path = ag_catalog, "$user", public;

SELECT * FROM cypher('music_genres', $$
CREATE (blues:Genre {name: 'Blues', type: 'base_genre'}),
(blues_description:Genre {name: 'blues_description', type: 'description'}),
(blues)-[:HAS_DESCRIPTION]->(blues_description)
SET blues.description = 'Blues is a music genre and musical form which was originated in 
the Deep South of the United States around the 1870s by African-Americans from roots in African musical traditions,
African-American work songs, and spirituals.'
$$) as (result agtype);

SELECT * FROM cypher('music_genres', $$
CREATE (acoustic_blues:Genre {name: 'Acoustic Blues', type: 'sub_genre'}),
(acoustic_blues_description:Genre {name: 'Acoustic Blues_description', type: 'description'}),
(blues)-[:HAS_DESCRIPTION]->(acoustic_blues_description)
SET acoustic_blues.description = 'Acoustic Blues is an unamplified style of blues guitar-playing that emerged in the early 1900s.
It is influenced by Delta blues, Piedmont blues, ragtime, and country blues. 
Characteristics include 12-bar blues progressions, varied accompaniment styles, and intricate fingerpicking.'
$$) as (result agtype);

SELECT * FROM cypher('music_genres', $$
MATCH (parent:Genre), (child: Genre)
WHERE parent.name = 'Blues' AND child.name = 'Acoustic Blues'
CREATE (parent)-[e:HAS_SUBGENRE]->(child)
RETURN e
$$) as (e agtype);
