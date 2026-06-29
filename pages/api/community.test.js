import { getClubsList, getRecommendData, normalizePostDetailResponse, processImages } from './community';

// Mock the request module
jest.mock('../../utils/request.js', () => ({
  request: jest.fn()
}));

import { request } from '../../utils/request.js';

describe('Community API', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('getClubsList should call request with correct parameters', async () => {
    const mockRes = { success: true, result: { records: [] } };
    request.mockResolvedValue(mockRes);

    const params = { page: 1, pageSize: 10 };
    const result = await getClubsList(params);

    expect(request).toHaveBeenCalledWith({
      url: '/api/clubs',
      method: 'GET',
      data: expect.objectContaining(params)
    });
    expect(result).toEqual(mockRes);
  });

  test('getRecommendData should call correct endpoint', async () => {
    const mockRes = { success: true, result: { clubs: [], posts: [] } };
    request.mockResolvedValue(mockRes);

    const result = await getRecommendData();

    expect(request).toHaveBeenCalledWith({
      url: '/api/community/recommend',
      method: 'GET'
    });
    expect(result).toEqual(mockRes);
  });

  test('normalizePostDetailResponse should unwrap result.post', () => {
    const payload = {
      success: true,
      result: {
        isLiked: true,
        isBookmarked: false,
        post: {
          id: 19,
          title: 't',
          content: 'c',
          images: '[]'
        }
      }
    };

    const post = normalizePostDetailResponse(payload);
    expect(post).toEqual({
      id: 19,
      title: 't',
      content: 'c',
      images: '[]',
      isLiked: true,
      isBookmarked: false
    });
  });

  test('processImages should parse malformed backtick images', () => {
    const raw = "[\" `http://a.com/1.webp\\` \",\" `http://a.com/2.webp\\` \"]";
    const result = processImages(raw);
    expect(result).toEqual(['http://a.com/1.webp', 'http://a.com/2.webp']);
  });
});
