#!/usr/bin/env python3
"""
Salmar AI Runner Script
-----------------------
This script provides a simple way to run the Salmar AI backend and connect to the frontend.
It also includes functionality to test if the system is working properly.
"""

import os
import sys
import time
import json
import signal
import argparse
import subprocess
import requests
from typing import Dict, List, Any, Optional
import webbrowser
from concurrent.futures import ThreadPoolExecutor

# Default configuration
DEFAULT_CONFIG = {
    "api_port": 8080,
    "nlp_service_port": 5001,
    "image_service_port": 5002,
    "code_service_port": 5003,
    "frontend_url": "http://localhost:3000",
    "api_url": "http://localhost:8080",
    "environment": "development"
}

# Global variables
running_processes = []

def parse_arguments():
    """Parse command line arguments"""
    parser = argparse.ArgumentParser(description="Salmar AI Runner")
    parser.add_argument("--test", action="store_true", help="Run tests to verify system functionality")
    parser.add_argument("--api-only", action="store_true", help="Run only the API service")
    parser.add_argument("--python-only", action="store_true", help="Run only the Python services")
    parser.add_argument("--no-browser", action="store_true", help="Don't open browser automatically")
    parser.add_argument("--env", choices=["development", "production"], default="development", 
                        help="Environment to run in (development or production)")
    return parser.parse_args()

def load_env_variables():
    """Load environment variables from .env file if it exists"""
    env_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), '.env')
    if os.path.exists(env_file):
        print("Loading environment variables from .env file")
        with open(env_file, 'r') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#'):
                    key, value = line.split('=', 1)
                    os.environ[key] = value

def run_command(command: List[str], cwd: Optional[str] = None) -> subprocess.Popen:
    """Run a command and return the process"""
    process = subprocess.Popen(
        command,
        cwd=cwd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=1,
        universal_newlines=True
    )
    running_processes.append(process)
    return process

def stream_output(process: subprocess.Popen, prefix: str):
    """Stream the output of a process with a prefix"""
    if process.stdout:
        for line in iter(process.stdout.readline, ''):
            if not line:
                break
            print(f"{prefix}: {line.strip()}")
    
    if process.stderr:
        for line in iter(process.stderr.readline, ''):
            if not line:
                break
            print(f"{prefix} ERROR: {line.strip()}")

def start_python_services():
    """Start the Python services"""
    print("Starting Python services...")
    
    # Start NLP service
    nlp_process = run_command(
        ["python", "nlp_service/app.py"],
        cwd=os.path.join(os.path.dirname(os.path.abspath(__file__)), "python")
    )
    
    # Start Image service
    image_process = run_command(
        ["python", "image_service/app.py"],
        cwd=os.path.join(os.path.dirname(os.path.abspath(__file__)), "python")
    )
    
    # Start Code service
    code_process = run_command(
        ["python", "code_service/app.py"],
        cwd=os.path.join(os.path.dirname(os.path.abspath(__file__)), "python")
    )
    
    # Stream output in separate threads
    with ThreadPoolExecutor(max_workers=3) as executor:
        executor.submit(stream_output, nlp_process, "NLP")
        executor.submit(stream_output, image_process, "IMAGE")
        executor.submit(stream_output, code_process, "CODE")
    
    return [nlp_process, image_process, code_process]

def check_go_dependencies():
    """Check and install required Go dependencies"""
    print("Checking Go dependencies...")
    
    # List of required Go dependencies
    required_deps = [
        "github.com/joho/godotenv",
        "github.com/gin-gonic/gin"
    ]
    
    # Check each dependency
    missing_deps = []
    for dep in required_deps:
        result = subprocess.run(
            ["go", "list", dep],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            cwd=os.path.dirname(os.path.abspath(__file__))
        )
        if result.returncode != 0:
            missing_deps.append(dep)
    
    # Install missing dependencies
    if missing_deps:
        print(f"Installing missing Go dependencies: {', '.join(missing_deps)}")
        subprocess.run(
            ["go", "get"] + missing_deps,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            cwd=os.path.dirname(os.path.abspath(__file__))
        )
        print("Dependencies installed successfully")

def start_api_service():
    """Start the Go API service"""
    print("Starting API service...")
    
    # Check Go dependencies first
    check_go_dependencies()
    
    # Set environment variables for the API service
    env = os.environ.copy()
    env["PORT"] = str(DEFAULT_CONFIG["api_port"])
    env["ENVIRONMENT"] = DEFAULT_CONFIG["environment"]
    env["PYTHON_SERVICES_HOST"] = "localhost"
    env["NLP_SERVICE_PORT"] = str(DEFAULT_CONFIG["nlp_service_port"])
    env["IMAGE_SERVICE_PORT"] = str(DEFAULT_CONFIG["image_service_port"])
    env["CODE_SERVICE_PORT"] = str(DEFAULT_CONFIG["code_service_port"])
    
    # Run the API service
    api_process = run_command(
        ["go", "run", "./cmd/api/main.go"],
        cwd=os.path.dirname(os.path.abspath(__file__))
    )
    
    # Stream output in a separate thread
    with ThreadPoolExecutor(max_workers=1) as executor:
        executor.submit(stream_output, api_process, "API")
    
    return api_process

def test_system():
    """Test if the system is functioning properly"""
    # Set test mode environment variable
    os.environ["SALMAR_TEST_MODE"] = "1"
    
    print("\n=== Testing Salmar AI System ===")
    
    # Test API service
    try:
        response = requests.get(f"{DEFAULT_CONFIG['api_url']}/health")
        if response.status_code == 200:
            print("✅ API service is running")
        else:
            print(f"❌ API service returned status code {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("❌ API service is not running")
    
    # Test NLP service
    try:
        response = requests.get(f"http://localhost:{DEFAULT_CONFIG['nlp_service_port']}/health")
        if response.status_code == 200:
            print("✅ NLP service is running")
        else:
            print(f"❌ NLP service returned status code {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("❌ NLP service is not running")
    
    # Test Image service
    try:
        response = requests.get(f"http://localhost:{DEFAULT_CONFIG['image_service_port']}/health")
        if response.status_code == 200:
            print("✅ Image service is running")
        else:
            print(f"❌ Image service returned status code {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("❌ Image service is not running")
    
    # Test Code service
    try:
        response = requests.get(f"http://localhost:{DEFAULT_CONFIG['code_service_port']}/health")
        if response.status_code == 200:
            print("✅ Code service is running")
        else:
            print(f"❌ Code service returned status code {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("❌ Code service is not running")
    
    # Test a simple request to the API
    try:
        test_request = {
            "text": "Hello, Salmar AI!",
            "user_id": "test_user",
            "conversation_id": "test_conversation"
        }
        response = requests.post(f"{DEFAULT_CONFIG['api_url']}/api/v1/process", json=test_request)
        if response.status_code == 200:
            print("✅ API request successful")
            print(f"Response: {response.json()}")
        else:
            print(f"❌ API request failed with status code {response.status_code}")
            print(f"Response: {response.text}")
    except requests.exceptions.ConnectionError:
        print("❌ API request failed - connection error")
    
    print("=== End of Tests ===\n")

def cleanup(signum=None, frame=None):
    """Clean up processes on exit"""
    print("\nShutting down Salmar AI...")
    for process in running_processes:
        if process.poll() is None:  # If process is still running
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
    print("All processes terminated")
    sys.exit(0)

def main():
    """Main function"""
    args = parse_arguments()
    
    # Set up signal handlers for graceful shutdown
    signal.signal(signal.SIGINT, cleanup)
    signal.signal(signal.SIGTERM, cleanup)
    
    # Load environment variables
    load_env_variables()
    
    # Set environment based on args
    DEFAULT_CONFIG["environment"] = args.env
    
    try:
        # Start services based on arguments
        if args.test:
            # Start all services and run tests
            if not args.python_only:
                start_api_service()
            if not args.api_only:
                start_python_services()
            
            # Wait for services to start
            print("Waiting for services to start...")
            time.sleep(5)
            
            # Run tests
            test_system()
            
        else:
            # Start services normally
            if not args.python_only:
                start_api_service()
            if not args.api_only:
                start_python_services()
            
            # Open browser if not disabled
            if not args.no_browser:
                print(f"Opening browser at {DEFAULT_CONFIG['frontend_url']}")
                time.sleep(3)  # Give services time to start
                webbrowser.open(DEFAULT_CONFIG['frontend_url'])
            
            print("\nSalmar AI is running!")
            print(f"API: http://localhost:{DEFAULT_CONFIG['api_port']}")
            print(f"NLP Service: http://localhost:{DEFAULT_CONFIG['nlp_service_port']}")
            print(f"Image Service: http://localhost:{DEFAULT_CONFIG['image_service_port']}")
            print(f"Code Service: http://localhost:{DEFAULT_CONFIG['code_service_port']}")
            print("\nPress Ctrl+C to stop")
            
            # Keep the script running
            while True:
                time.sleep(1)
                
    except KeyboardInterrupt:
        pass
    finally:
        cleanup()

if __name__ == "__main__":
    main()