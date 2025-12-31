# SSH-enabled container for jump host testing
FROM ubuntu:22.04

# Install OpenSSH server and other utilities
RUN apt-get update && \
    apt-get install -y openssh-server sudo curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create SSH directory
RUN mkdir -p /var/run/sshd

# Create test user with sudo privileges
RUN useradd -m -s /bin/bash testuser && \
    echo "testuser:testpass" | chpasswd && \
    usermod -aG sudo testuser

# Configure SSH
RUN mkdir -p /home/testuser/.ssh && \
    chmod 700 /home/testuser/.ssh && \
    chown testuser:testuser /home/testuser/.ssh

# Enable SSH key authentication and disable password auth for security
RUN sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config && \
    sed -i 's/UsePAM yes/UsePAM no/' /etc/ssh/sshd_config

# Expose SSH port
EXPOSE 22

# Start SSH service
CMD ["/usr/sbin/sshd", "-D"]
