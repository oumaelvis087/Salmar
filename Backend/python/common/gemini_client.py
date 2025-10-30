"""
Gemini API Client for Salmar AI
This module provides a client for Google's Gemini API to handle multimodal reasoning.
"""

import os
import base64
import json
from typing import Dict, List, Any, Optional, Union
import google.generativeai as genai
from PIL import Image
import io

# Import configuration
from common.config import GOOGLE_API_KEY, GEMINI_TEXT_MODEL, GEMINI_VISION_MODEL, GEMINI_TEMPERATURE

class GeminiClient:
    """Client for interacting with Google's Gemini API"""
    
    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize the Gemini client
        
        Args:
            api_key: Google AI Studio API key (defaults to GOOGLE_API_KEY from config)
        """
        self.api_key = api_key or GOOGLE_API_KEY or "dummy_key_for_testing"
        
        # For testing purposes, we'll skip actual API configuration if in test mode
        if os.getenv("SALMAR_TEST_MODE") == "1":
            print("Running in test mode - skipping Gemini API configuration")
        else:
            # Configure the Gemini API
            genai.configure(api_key=self.api_key)
        
        # Available models from configuration
        self.text_model = GEMINI_TEXT_MODEL  # For text and code
        self.vision_model = GEMINI_VISION_MODEL  # For multimodal (text + images)
        
    def chat(self, messages: List[Dict[str, str]], temperature: float = 0.7) -> Dict[str, Any]:
        """
        Generate a chat response using Gemini
        
        Args:
            messages: List of message dictionaries with 'role' and 'content' keys
            temperature: Controls randomness (0.0 to 1.0)
            
        Returns:
            Dictionary containing the response and metadata
        """
        # Convert to Gemini format
        gemini_messages = []
        for msg in messages:
            role = "user" if msg["role"] == "user" else "model"
            gemini_messages.append({"role": role, "parts": [{"text": msg["content"]}]})
        
        # Create chat session
        model = genai.models.GenerativeModel(self.text_model, generation_config={"temperature": temperature})
        chat = model.start_chat(history=gemini_messages)
        
        # Generate response
        response = chat.send_message("")
        
        # Format response
        return {
            "text": response.text,
            "model_info": {
                "model": self.text_model,
                "temperature": temperature,
                "usage": {
                    "prompt_tokens": response.usage_metadata.prompt_token_count if hasattr(response, 'usage_metadata') else 0,
                    "completion_tokens": response.usage_metadata.candidates_token_count if hasattr(response, 'usage_metadata') else 0
                }
            }
        }
    
    def generate_code(self, prompt: str, temperature: float = 0.2) -> Dict[str, Any]:
        """
        Generate code using Gemini
        
        Args:
            prompt: The code generation prompt
            temperature: Controls randomness (0.0 to 1.0)
            
        Returns:
            Dictionary containing the generated code and metadata
        """
        # Add code-specific instructions
        enhanced_prompt = f"Generate code for the following task. Return only the code without explanations unless specifically requested:\n\n{prompt}"
        
        # Create model
        model = genai.models.GenerativeModel(self.text_model, generation_config={"temperature": temperature})
        
        # Generate response
        response = model.generate_content(enhanced_prompt)
        
        # Extract code blocks from response
        code_blocks = self._extract_code_blocks(response.text)
        
        return {
            "text": response.text,
            "code_blocks": code_blocks,
            "model_info": {
                "model": self.text_model,
                "temperature": temperature,
                "usage": {
                    "prompt_tokens": response.usage_metadata.prompt_token_count if hasattr(response, 'usage_metadata') else 0,
                    "completion_tokens": response.usage_metadata.candidates_token_count if hasattr(response, 'usage_metadata') else 0
                }
            }
        }
    
    def analyze_image(self, image_data: Union[str, bytes], prompt: str) -> Dict[str, Any]:
        """
        Analyze an image using Gemini Vision
        
        Args:
            image_data: Base64 encoded image string or raw image bytes
            prompt: The analysis prompt
            
        Returns:
            Dictionary containing the analysis and metadata
        """
        # Process image data
        if isinstance(image_data, str):
            # Assume base64 encoded
            image_bytes = base64.b64decode(image_data)
        else:
            # Raw bytes
            image_bytes = image_data
            
        # Load image
        image = Image.open(io.BytesIO(image_bytes))
        
        # Create model
        model = genai.models.GenerativeModel(self.vision_model)
        
        # Generate response
        response = model.generate_content([prompt, image])
        
        return {
            "text": response.text,
            "model_info": {
                "model": self.vision_model,
                "usage": {
                    "prompt_tokens": response.usage_metadata.prompt_token_count if hasattr(response, 'usage_metadata') else 0,
                    "completion_tokens": response.usage_metadata.candidates_token_count if hasattr(response, 'usage_metadata') else 0
                }
            }
        }
    
    def multimodal_reasoning(self, text: str, images: Optional[List[Union[str, bytes]]] = None) -> Dict[str, Any]:
        """
        Process a multimodal request with text and optional images
        
        Args:
            text: The text prompt
            images: List of base64 encoded image strings or raw image bytes
            
        Returns:
            Dictionary containing the response and metadata
        """
        # Create content parts
        parts = [text]
        
        # Add images if provided
        if images:
            for img_data in images:
                if isinstance(img_data, str):
                    # Assume base64 encoded
                    image_bytes = base64.b64decode(img_data)
                else:
                    # Raw bytes
                    image_bytes = img_data
                
                # Load image
                img = Image.open(io.BytesIO(image_bytes))
                parts.append(img)
        
        # Create model
        model = genai.models.GenerativeModel(self.vision_model)
        
        # Generate response
        response = model.generate_content(parts)
        
        return {
            "text": response.text,
            "model_info": {
                "model": self.vision_model,
                "usage": {
                    "prompt_tokens": response.usage_metadata.prompt_token_count if hasattr(response, 'usage_metadata') else 0,
                    "completion_tokens": response.usage_metadata.candidates_token_count if hasattr(response, 'usage_metadata') else 0
                }
            }
        }
    
    def _extract_code_blocks(self, text: str) -> List[Dict[str, str]]:
        """Extract code blocks from markdown text"""
        code_blocks = []
        lines = text.split('\n')
        in_code_block = False
        current_block = {"language": "", "code": ""}
        
        for line in lines:
            if line.startswith('```'):
                if in_code_block:
                    # End of code block
                    in_code_block = False
                    if current_block["code"]:
                        code_blocks.append(current_block)
                    current_block = {"language": "", "code": ""}
                else:
                    # Start of code block
                    in_code_block = True
                    language = line[3:].strip()
                    current_block["language"] = language
            elif in_code_block:
                current_block["code"] += line + '\n'
        
        return code_blocks