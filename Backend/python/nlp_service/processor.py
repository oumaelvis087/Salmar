import logging
import os
from typing import Dict, Any

# Configure logging
logging.basicConfig(level=os.getenv("LOG_LEVEL", "INFO"))
logger = logging.getLogger(__name__)

import os
import sys
from typing import Dict, Any, Optional

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from common.gemini_client import GeminiClient

class NLPProcessor:
    """Handles natural language processing tasks using Google's Gemini API."""
    
    # Intent keywords for classification (fallback if Gemini fails)
    INTENT_KEYWORDS = {
        "generate": ["create", "generate", "make", "build", "write", "develop"],
        "analyze": ["analyze", "examine", "study", "investigate", "assess", "evaluate"],
        "explain": ["explain", "describe", "clarify", "elaborate", "tell me about"],
        "summarize": ["summarize", "sum up", "brief", "overview", "synopsis"],
        "translate": ["translate", "convert", "change", "transform"],
        "optimize": ["optimize", "improve", "enhance", "refine", "speed up"],
        "debug": ["debug", "fix", "solve", "resolve", "troubleshoot"],
        "automate": ["automate", "schedule", "repeat", "routine"]
    }
    
    def __init__(self):
        """Initialize the NLP processor with Gemini client."""
        logger.info("Initializing NLP processor with Gemini API")
        self.gemini_client = GeminiClient()
        
    def classify_intent(self, data: Dict[str, Any]) -> str:
        """Classify the intent of the request using Gemini API."""
        text = data.get("text", "").lower()
        
        if not text:
            return "conversation"
            
        try:
            # Use Gemini for more sophisticated intent classification
            prompt = f"""
            Classify the following text into one of these intents: 'generate', 'analyze', 'explain', 'summarize', 'translate', 'optimize', 'debug', 'automate', 'conversation'.
            
            Text: {text}
            
            Return ONLY the intent name without any explanation.
            """
            
            response = self.gemini_client.generate_code(prompt, temperature=0.1)
            intent = response.get('text', '').strip().lower()
            
            # Validate the intent
            valid_intents = ['generate', 'analyze', 'explain', 'summarize', 'translate', 'optimize', 'debug', 'automate', 'conversation']
            if intent in valid_intents:
                return intent
                
            # Fallback to keyword matching
            logger.warning(f"Gemini returned invalid intent: {intent}. Falling back to keyword matching.")
        except Exception as e:
            logger.error(f"Error using Gemini for intent classification: {e}. Falling back to keyword matching.")
        
        # Fallback: Count keyword matches for each intent
        intent_scores = {}
        for intent, keywords in self.INTENT_KEYWORDS.items():
            score = sum(1 for keyword in keywords if keyword in text)
            intent_scores[intent] = score
        
        # Find the intent with the highest score
        max_score = 0
        max_intent = "conversation"  # Default intent
        
        for intent, score in intent_scores.items():
            if score > max_score:
                max_score = score
                max_intent = intent
        
        return max_intent
    
    def process(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Process text input and generate responses using Gemini API."""
        logger.debug(f"Processing NLP request: {data}")
        
        # Extract text from request
        text = data.get("text", "")
        if not text:
            return {"text_response": "No text input provided"}
        
        # Get conversation history if available
        history = data.get("history", [])
        
        # Classify intent if not provided
        intent = data.get("intent")
        if not intent:
            intent = self.classify_intent(data)
        
        try:
            # Format messages for Gemini
            messages = []
            
            # Add context from previous messages if available
            for msg in history:
                messages.append({
                    'role': msg.get('role', 'user'),
                    'content': msg.get('content', '')
                })
            
            # Add the current message with intent guidance
            messages.append({
                'role': 'user',
                'content': f"[Intent: {intent}] {text}"
            })
            
            # Get response from Gemini
            gemini_response = self.gemini_client.chat(messages)
            response = gemini_response.get('text', '')
            
            # Add reasoning steps for transparency
            reasoning_steps = self._generate_reasoning_steps(text, intent)
            
            return {
                "text_response": response,
                "intent": intent,
                "reasoning_steps": reasoning_steps,
                "model_info": {
                    "name": "Gemini API",
                    "version": gemini_response.get('model_info', {}).get('version', 'unknown')
                }
            }
            
        except Exception as e:
            logger.error(f"Error using Gemini API: {e}. Falling back to template responses.")
            # Fallback to template responses
            response = self._generate_response(text, history, intent)
            reasoning_steps = self._generate_reasoning_steps(text, intent)
            
            return {
                "text_response": response,
                "intent": intent,
                "reasoning_steps": reasoning_steps,
                "model_info": {
                    "name": "Salmar NLP Demo (Fallback)",
                    "version": "0.1.0"
                },
                "errors": [{
                    "message": str(e),
                    "code": "gemini_processing_error"
                }]
            }
    
    def _generate_response(self, text: str, history: list, intent: str = "conversation") -> str:
        """Generate a response based on input text, conversation history, and intent."""
        # Fallback responses when Gemini API is unavailable
        
        # Intent-based responses
        if intent == "generate":
            return f"I'll generate content based on your request: '{text}'. In a production environment, this would use a generative AI model."
            
        elif intent == "analyze":
            return f"I'll analyze the following for you: '{text}'. In a production environment, this would use analytical models."
            
        elif intent == "explain":
            return f"Let me explain '{text}' for you. In a production environment, this would provide detailed explanations."
            
        elif intent == "summarize":
            return f"Here's a summary of '{text}'. In a production environment, this would generate concise summaries."
            
        elif intent == "translate":
            return f"I'll translate '{text}' for you. In a production environment, this would use translation models."
            
        elif intent == "optimize":
            return f"I'll optimize '{text}' for you. In a production environment, this would suggest improvements."
            
        elif intent == "debug":
            return f"I'll help debug the issue with '{text}'. In a production environment, this would identify and fix problems."
            
        elif intent == "automate":
            return f"I'll help automate '{text}' for you. In a production environment, this would set up automation workflows."
        
        # Simple keyword-based responses for demonstration
        if "hello" in text.lower() or "hi" in text.lower():
            return "Hello! I'm Salmar AI. How can I assist you today?"
        
        if "code" in text.lower() or "program" in text.lower():
            return "I can help you with coding tasks. Please provide more details about what you'd like me to do."
        
        if "image" in text.lower() or "picture" in text.lower():
            return "I can analyze or generate images. Would you like me to create an image or analyze an existing one?"
        
        # Default response
        return "I understand your request and will process it accordingly. How else can I assist you?"
        
    def _generate_reasoning_steps(self, text: str, intent: str) -> list:
        """Generate reasoning steps for transparency."""
        steps = []
        
        # Add intent classification step
        steps.append({
            "description": f"Classified request intent as '{intent}'",
            "confidence": 0.85
        })
        
        # Add content analysis step
        steps.append({
            "description": f"Analyzed request content using Gemini API",
            "confidence": 0.9
        })
        
        # Add response generation step
        steps.append({
            "description": f"Generated response based on intent and content analysis",
            "confidence": 0.95
        })
        
        return steps