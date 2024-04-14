// TODO: migrate from jest to bun's integrated testing
import axios from "axios";
import MockAdapter from "axios-mock-adapter";
import { castStore } from "./cast";

// Create an instance of axios-mock-adapter
const mock = new MockAdapter(axios);

describe("castStore", () => {
  let mockAxios: MockAdapter;

  beforeAll(() => {
    mockAxios = new MockAdapter(axios);
  });

  afterAll(() => {
    mockAxios.restore();
  });

  beforeEach(() => {
    // Reset the store state before each test
    castStore.set({ ID: 0, actors: [], directors: [] });
  });

  it("should fetch and update cast data correctly", async () => {
    // TODO: populate the test media object in the DB, adjust expectations accordingly
    // then create a migration with the INSERT statements for the test data
    const testMediaId = "6931b501-f1bd-4dc8-9280-f86d4bb81978";
    const mockResponseData = {
      data: {
        ID: 1,
        actors: [{ name: "Actor 1" }, { name: "Actor 2" }],
        directors: [{ name: "Director 1" }],
      },
    };

    // Mock the Axios GET request with the expected URL and response data
    mockAxios.onGet(`/api/media/${testMediaId}/cast/`).reply(200, mockResponseData);

    // Call the getCast function with the test media ID and await the result
    await castStore.getCast(testMediaId);

    // Use a setTimeout to wait for the store to update
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Access the store state
    const storeState = castStore.subscribe((value) => {
      // Check if the store state was updated correctly
      expect(value.ID).toBe(1);
      expect(value.actors).toEqual([{ name: "Actor 1" }, { name: "Actor 2" }]);
      expect(value.directors).toEqual([{ name: "Director 1" }]);
    });

    // Clean up the subscription
    storeState();
  });

  it("should handle error when fetching cast data", async () => {
    const testMediaId = "6931b501-f1bd-4dc8-9280-f86d4bb81978";

    // Mock a failed Axios GET request with a 500 internal server error
    mockAxios.onGet(`/api/media/${testMediaId}/cast/`).reply(500);

    // Call the getCast function with the test media ID and await the result
    await castStore.getCast(testMediaId);

    // Use a setTimeout to wait for the store to update
    await new Promise((resolve) => setTimeout(resolve, 0));

    // Access the store state
    const storeState = castStore.subscribe((value) => {
      // Check if the store state remains in its initial state (error handling)
      expect(value.ID).toBe(0);
      expect(value.actors).toEqual([]);
      expect(value.directors).toEqual([]);
    });

    // Clean up the subscription
    storeState();
  });
});
