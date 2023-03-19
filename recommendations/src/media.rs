use ndarray::{Array2, Array1};
use serde::Deserialize;
use serde_derive::Deserialize;
use std::collections::HashSet;

#[derive(Debug, Deserialize)]
pub struct Media {
    pub id: u32,
    pub media_type: String,
    pub title: String,
    pub genres: Vec<String>,
    pub year: u32,
    pub people: Vec<String>,
    pub influences: Vec<String>,
    pub countries: Vec<String>,
    pub languages: Vec<String>,
    pub length: u32,
}

pub fn generate_feature_vectors(media_list: &[Media]) -> (Array2<f64>, Vec<u32>, HashSet<String>) {
    let mut unique_genres = HashSet::new();
    let mut unique_people = HashSet::new();
    let mut unique_influences = HashSet::new();
    let mut unique_countries = HashSet::new();
    let mut unique_languages = HashSet::new();

    for media in media_list {
        unique_genres.extend(media.genres.iter().cloned());
        unique_people.extend(media.people.iter().cloned());
        unique_influences.extend(media.influences.iter().cloned());
        unique_countries.extend(media.countries.iter().cloned());
        unique_languages.extend(media.languages.iter().cloned());
    }

    let feature_count = unique_genres.len()
        + unique_people.len()
        + unique_influences.len()
        + unique_countries.len()
        + unique_languages.len();
    let mut feature_vectors = Array2::<f64>::zeros((media_list.len(), feature_count));
    let mut id_list = Vec::new();

    let mut index = 0;
    for (row, media) in media_list.iter().enumerate() {
        id_list.push(media.id);
        
        for genre in &media.genres {
            let feature_index = unique_genres.get_index_of(genre).unwrap();
            feature_vectors[[row, index + feature_index]] = 1.0;
        }
        index += unique_genres.len();

        for person in &media.people {
            let feature_index = unique_people.get_index_of(person).unwrap();
            feature_vectors[[row, index + feature_index]] = 1.0;
        }
        index += unique_people.len();

        for influence in &media.influences {
            let feature_index = unique_influences.get_index_of(influence).unwrap();
            feature_vectors[[row, index + feature_index]] = 1.0;
        }
        index += unique_influences.len();

        for country in &media.countries {
            let feature_index = unique_countries.get_index_of(country).unwrap();
            feature_vectors[[row, index + feature_index]] = 1.0;
        }
        index += unique_countries.len();

        for language in &media.languages {
            let feature_index = unique_languages.get_index_of(language).unwrap();
            feature_vectors[[row, index + feature_index]] = 1.0;
        }
        index = 0;
    }

    (feature_vectors, id_list, unique_genres)
}

trait HashSetIndex<T: PartialEq> {
    fn get_index_of(&self, value: &T) -> Option<usize>;
}

impl<T: PartialEq> HashSetIndex<T> for HashSet<T> {
    fn get_index_of(&self, value: &T) -> Option<usize> {
        self.iter().position(|x| x == value)
    }
}
