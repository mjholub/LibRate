use ndarray::Array2;
use std::fs::File;
use std::io::BufReader;
use std::error::Error;
use crate::media::{Media, generate_feature_vectors};

mod media;

fn main() -> Result<(), Box<dyn Error>> {
    let media_data_file = File::open("media_data.csv")?;
    let media_data_reader = BufReader::new(media_data_file);

    let media_list: Vec<Media> = csv::Reader::from_reader(media_data_reader)
        .deserialize()
        .collect::<Result<Vec<Media>, csv::Error>>()?;

    let (feature_vectors, id_list, unique_genres) = generate_feature_vectors(&media_list);

    println!("Feature vectors shape: {:?}", feature_vectors.dim());
    println!("Media IDs: {:?}", id_list);
    println!("Unique genres: {:?}", unique_genres);

    Ok(())
}
