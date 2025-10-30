#!/usr/bin/env python3
"""
Service launcher for Salmar AI Python services.
This script starts all the Python microservices required by the Salmar AI system.
"""

import os
import sys
import subprocess
import time
import signal
import argparse
from typing import List, Dict

# Service definitions
SERVICES = {
    "nlp": {
        "dir": "nlp_service",
        "port": 5001,
        "env": {"NLP_SERVICE_PORT": "5001", "SERVICE_NAME": "nlp"}
    },
    "image": {
        "dir": "image_service",
        "port": 5002,
        "env": {"IMAGE_SERVICE_PORT": "5002", "SERVICE_NAME": "image"}
    },
    "code": {
        "dir": "code_service",
        "port": 5003,
        "env": {"CODE_SERVICE_PORT": "5003", "SERVICE_NAME": "code"}
    }
}

# Global process tracking
processes: Dict[str, subprocess.Popen] = {}

def start_service(service_name: str, debug: bool = False) -> None:
    """Start a specific service."""
    if service_name not in SERVICES:
        print(f"Unknown service: {service_name}")
        return
    
    service = SERVICES[service_name]
    service_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), service["dir"])
    
    # Prepare environment
    env = os.environ.copy()
    env.update(service["env"])
    if debug:
        env["DEBUG"] = "true"
    
    # Start the service
    cmd = [sys.executable, "app.py"]
    print(f"Starting {service_name} service on port {service['port']}...")
    
    process = subprocess.Popen(
        cmd,
        cwd=service_dir,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        universal_newlines=True
    )
    
    processes[service_name] = process
    print(f"{service_name} service started with PID {process.pid}")

def start_all_services(debug: bool = False) -> None:
    """Start all services."""
    for service_name in SERVICES:
        start_service(service_name, debug)

def stop_service(service_name: str) -> None:
    """Stop a specific service."""
    if service_name in processes:
        print(f"Stopping {service_name} service...")
        processes[service_name].terminate()
        processes[service_name].wait(timeout=5)
        del processes[service_name]
        print(f"{service_name} service stopped")

def stop_all_services() -> None:
    """Stop all running services."""
    for service_name in list(processes.keys()):
        stop_service(service_name)

def signal_handler(sig, frame) -> None:
    """Handle termination signals."""
    print("Shutting down all services...")
    stop_all_services()
    sys.exit(0)

def main() -> None:
    """Main entry point."""
    parser = argparse.ArgumentParser(description="Salmar AI Python Services Launcher")
    parser.add_argument("--debug", action="store_true", help="Run services in debug mode")
    parser.add_argument("--service", type=str, help="Start a specific service only")
    args = parser.parse_args()
    
    # Register signal handlers
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        if args.service:
            start_service(args.service, args.debug)
        else:
            start_all_services(args.debug)
        
        # Keep the script running
        print("All services started. Press Ctrl+C to stop.")
        while True:
            time.sleep(1)
            
            # Check if any process has terminated
            for service_name, process in list(processes.items()):
                if process.poll() is not None:
                    print(f"{service_name} service terminated unexpectedly with code {process.returncode}")
                    # Restart the service
                    start_service(service_name, args.debug)
    
    except KeyboardInterrupt:
        print("Interrupted by user")
    finally:
        stop_all_services()

if __name__ == "__main__":
    main()