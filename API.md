# Review Guess API

Simple REST API for fetching Letterboxd reviews from one or multiple users.

## Endpoints

### Health Check

```
GET /health
```

Verifies the API is running.

**Response:** 200 OK
```json
{
  "success": true,
  "data": "Review Guess API v1.0"
}
```

---

### Get Reviews

```
GET /api/reviews?username={username}[&username={username}]...
```

Fetches reviews from specified user(s).

**Parameters:**
- `username` (required, repeatable): Letterboxd username(s)

**Query Examples:**
- Single user: `GET /api/reviews?username=alice`
- Multiple users: `GET /api/reviews?username=alice&username=bob&username=charlie`

**Success Response:** 200 OK
```json
{
  "success": true,
  "data": {
    "count": 25,
    "reviews": [
      {
        "author": "alice",
        "title": "The Shawshank Redemption",
        "slug": "the-shawshank-redemption-1994",
        "content": "A masterpiece of cinema...",
        "rating": 5,
        "liked": true,
        "spoilers": false
      },
      {
        "author": "bob",
        "title": "Inception",
        "slug": "inception-2010",
        "content": "Mind-bending and brilliant...",
        "rating": 4,
        "liked": false,
        "spoilers": false
      }
    ]
  }
}
```

**Error Response:** 400 Bad Request
```json
{
  "success": false,
  "error": "username parameter is required"
}
```

## Review Object

| Field | Type | Description |
|-------|------|-------------|
| `author` | string | Letterboxd username |
| `title` | string | Film title |
| `slug` | string | Unique film slug (for linking) |
| `content` | string | Review text |
| `rating` | int | Rating 0-5 (0 = watched, 1-5 = stars) |
| `liked` | boolean | Whether the review is liked |
| `spoilers` | boolean | Whether marked as spoilers |

## HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad request (missing username parameter) |
| 500 | Server error (scraping failed) |

## Examples

### Using curl

```bash
# Single user
curl "http://localhost:8080/api/reviews?username=alice"

# Multiple users
curl "http://localhost:8080/api/reviews?username=alice&username=bob"
```

### Using fetch (JavaScript)

```javascript
const usernames = ['alice', 'bob'];
const params = usernames.map(u => `username=${u}`).join('&');

fetch(`http://localhost:8080/api/reviews?${params}`)
  .then(res => res.json())
  .then(data => {
    if (data.success) {
      console.log(`Got ${data.data.count} reviews`);
      data.data.reviews.forEach(review => {
        console.log(`${review.author}: ${review.title}`);
      });
    }
  });
```

### Using Python

```python
import requests

usernames = ['alice', 'bob']
params = {'username': usernames}

response = requests.get('http://localhost:8080/api/reviews', params=params)
data = response.json()

if data['success']:
    print(f"Got {data['data']['count']} reviews")
    for review in data['data']['reviews']:
        print(f"{review['author']}: {review['title']}")
```

## Rate Limiting

The scraper includes built-in rate limiting:
- 3 seconds delay between pages
- 2 seconds random delay
- Respects Letterboxd's terms of service

## Notes

- Reviews are scraped in real-time from Letterboxd
- Results are not cached
- Multiple requests for the same user will scrape again
- Minimum review content: empty reviews are skipped

