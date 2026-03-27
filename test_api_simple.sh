#!/bin/bash

# 🎬 Script de test simple pour l'API Review Guess

API="http://localhost:8080"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🧪 Testing Review Guess API${NC}\n"

# 1. Health Check
echo -e "${GREEN}1. 🏥 Health Check${NC}"
curl -s $API/health | python3 -m json.tool
echo -e "\n"

# 2. Fetch Reviews
echo -e "${GREEN}2. 📚 Fetching reviews from 66sceptre${NC}"
curl -s "$API/api/reviews?username=66sceptre" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    print(f\"✓ Found {data['data']['count']} reviews\")
    print(f\"✓ First review: '{data['data']['reviews'][0]['title']}'\")"
echo -e "\n"

# 3. Start Game
echo -e "${GREEN}3. 🎮 Starting game${NC}"
GAME=$(curl -s -X POST $API/api/game/start \
  -H "Content-Type: application/json" \
  -d '{"usernames": ["66sceptre"], "question_count": 3}')

echo "$GAME" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    print(f\"✓ Game started\")
    q = data['data']['current_question']
    print(f\"✓ Question 1/{data['data']['total_questions']}: '{q['review']['title']}'\")
    print(f\"✓ Difficulty: {q['difficulty']:.1f}\")"
echo -e "\n"

# 4. Get Current Question
echo -e "${GREEN}4. ❓ Getting current question${NC}"
curl -s $API/api/game/question | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    q = data['data']
    print(f\"✓ Question {q['index']}/{q['total']}\")
    print(f\"✓ Review: \\\"{q['review']['content'][:80]}...\\\"\")
    print(f\"✓ Rating: {'⭐' * q['review']['rating']}\")"
echo -e "\n"

# 5. Submit Correct Answer
echo -e "${GREEN}5. ✅ Submitting CORRECT answer${NC}"
ANSWER=$(curl -s -X POST $API/api/game/answer \
  -H "Content-Type: application/json" \
  -d '{"guessed_author": "66sceptre", "guessed_film": "project-hail-mary"}')

echo "$ANSWER" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    d = data['data']
    print(f\"✓ Result: {'🎯 CORRECT!' if d['correct'] else '❌ Wrong'}\")
    print(f\"✓ Points: {d['points']}\")
    print(f\"✓ Score: {d['current_score']}\")"
echo -e "\n"

# 6. Next Question
echo -e "${GREEN}6. ❓ Getting next question${NC}"
curl -s $API/api/game/question | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    q = data['data']
    print(f\"✓ Question {q['index']}/{q['total']}\")"
echo -e "\n"

# 7. Submit Wrong Answer
echo -e "${GREEN}7. ❌ Submitting WRONG answer${NC}"
WRONG=$(curl -s -X POST $API/api/game/answer \
  -H "Content-Type: application/json" \
  -d '{"guessed_author": "wrong", "guessed_film": "wrong"}')

echo "$WRONG" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    d = data['data']
    print(f\"✓ Result: {'🎯 CORRECT!' if d['correct'] else '❌ Wrong'}\")
    print(f\"✓ Points: {d['points']}\")
    print(f\"✓ Score: {d['current_score']}\")"
echo -e "\n"

# 8. Check Score
echo -e "${GREEN}8. 📊 Checking current score${NC}"
curl -s $API/api/game/score | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    d = data['data']
    print(f\"✓ Score: {d['current_score']}/{d['total_questions']*100}\")
    print(f\"✓ Progress: {d['answered']}/{d['total_questions']}\")"
echo -e "\n"

# 9. Submit Last Answer
echo -e "${GREEN}9. 🔚 Submitting answer to last question${NC}"
LAST=$(curl -s -X POST $API/api/game/answer \
  -H "Content-Type: application/json" \
  -d '{"guessed_author": "66sceptre", "guessed_film": "wrong"}')

echo "$LAST" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    d = data['data']
    print(f\"✓ Game Over: {d['is_game_over']}\")
    print(f\"✓ Final Score: {d['current_score']}\")"
echo -e "\n"

# 10. Get Final Results
echo -e "${GREEN}10. 🏆 Getting final results${NC}"
curl -s $API/api/game/results | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data.get('success'):
    d = data['data']
    print(f\"✓ Final Score: {d['score']}/{d['total_points']}\")
    print(f\"✓ Percentage: {d['percentage']}%\")
    print(f\"✓ Grade: {d['grade']}\")
    print(f\"✓ Answered: {d['answered']}/{d['total']}\")"
echo -e "\n"

echo -e "${YELLOW}✅ Test complete!${NC}"
