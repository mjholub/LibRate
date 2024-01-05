import json
import os

def generate_cypher_query(base_genre, genre_data):
    cypher_queries = []

    # CREATE statement for base_genre node
    cypher_queries.append(f"CREATE ({base_genre}:Genre {{name: '{base_genre}', type: 'base_genre'}})")

    # CREATE statement for base_genre_description node
    cypher_queries.append(f"CREATE ({base_genre}_description:Genre {{name: '{base_genre}_description', type: 'description'}})")

    # Relationship between base_genre and base_genre_description
    cypher_queries.append(f"CREATE ({base_genre})-[:HAS_DESCRIPTION]->({base_genre}_description)")

    for genre in genre_data:
    # CREATE statement for each genre node
        cypher_query = f"CREATE ({genre['name']}:Genre {{name: '{genre['name']}'}})"
        
        if 'description' in genre:
            cypher_query += f"\nSET {genre['name']}.description = '{genre['description']}'"

        # Relationship between base_genre and genre
        cypher_query += f"\nCREATE ({base_genre})-[:HAS_CHILD]->({genre['name']})"
        
        # Relationship between genre and base_genre_description
        cypher_query += f"\nCREATE ({genre['name']})-[:HAS_DESCRIPTION]->({base_genre}_description)"

        if 'children' in genre and genre['children']:
            for child in genre['children']:
                # CREATE statement for each child node
                cypher_query += f"\nCREATE ({child['name']}:Genre {{name: '{child['name']}'}})"
                if 'description' in child:
                    cypher_query += f"\nSET {child['name']}.description = '{child['description']}'"

                # Relationship between genre and child
                cypher_query += f"\nCREATE ({genre['name']})-[:HAS_CHILD]->({child['name']})"
                
                # Relationship between child and base_genre_description
                cypher_query += f"\nCREATE ({child['name']})-[:HAS_DESCRIPTION]->({base_genre}_description)"

        cypher_queries.append(cypher_query)

    return '\n\n'.join(cypher_queries)

def main():
    # Specify the directory containing your JSON files
    json_directory = '.'

    # Iterate through JSON files in the directory
    for filename in os.listdir(json_directory):
        if filename.endswith('.json'):
            base_genre = os.path.splitext(filename)[0]
            
            # Load genre data from JSON file
            with open(os.path.join(json_directory, filename), 'r') as file:
                genre_data = json.load(file)

            # Generate Cypher queries for each JSON file
            cypher_queries = generate_cypher_query(base_genre, genre_data)

            # Print or save the queries as needed
            print(cypher_queries)

if __name__ == "__main__":
    main()
