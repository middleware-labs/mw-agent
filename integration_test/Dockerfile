FROM ubuntu:latest

# Use the official Python image as a base image
FROM python:3.8

# Set the working directory in the container
WORKDIR /app

# Copy the requirements file into the container at /app
COPY requirements.txt /app/

# Install any dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the current directory contents into the container at /app
COPY . /app/

RUN cd /app/

# Specify the command to run on container start
CMD ["python3", "integration_test.py"]
