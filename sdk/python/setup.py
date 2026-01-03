from setuptools import setup, find_packages

setup(
    name="engram-sdk",
    version="0.1.0",
    packages=find_packages(),
    install_requires=[
        "grpcio>=1.76.0",
        "protobuf>=6.33.2",
    ],
    author="Zynta",
    description="Durable AI Agent Engine SDK for Python",
    python_requires=">=3.9",
)
