use tonic::{transport::Server, Request, Response, Status};

pub mod recommendation_service {
    tonic::include_proto!("recommendation"); // The string specified here must match the proto package name
}

use recommendation_service::recommendation_service_server::{RecommendationService, RecommendationServiceServer};
use recommendation_service::{GetRecommendationsRequest, GetRecommendationsResponse};

#[derive(Debug, Default)]
pub struct LibRateRecSvc {}

#[tonic::async_trait]
impl RecommendationService for LibRateRecSvc {
    async fn get_recommendations(
        &self,
        request: Request<GetRecommendationsRequest>, // Accept request of type GetRecommendationsRequest
    ) -> Result<Response<GetRecommendationsResponse>, Status> { // Return an instance of type GetRecommendationsResponse
        
        println!("Got a request: {:?}", request);

        let reply = GetRecommendationsResponse {
            // Insert your code logic here
        };

        Ok(Response::new(reply)) // Send back our formatted greeting
    }
}
