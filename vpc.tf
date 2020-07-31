resource "aws_vpc" "app" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = var.vpc_name
  }
}

resource "aws_eip" "ngw" {
  vpc = true
  tags = {
    Name = "${var.vpc_name}-ngw"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.app.id
  tags = {
    Name = var.vpc_name
  }
}

resource "aws_default_security_group" "default" {
  vpc_id = aws_vpc.app.id
  tags = {
    Name = "${var.vpc_name}-default"
  }
}

resource "aws_security_group" "egress-only" {
  name        = "egress_only"
  description = "Allow all outbound traffic, no inbound"
  vpc_id      = aws_vpc.app.id
  tags = {
    Name = "${var.vpc_name}-egress-only"
  }
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}


# public
resource "aws_subnet" "public" {
  vpc_id     = aws_vpc.app.id
  cidr_block = "10.0.0.0/24"
  tags = {
    Name = "${var.vpc_name}-public"
  }
}

resource "aws_nat_gateway" "ngw" {
  subnet_id     = aws_subnet.public.id
  allocation_id = aws_eip.ngw.id
  tags = {
    Name = var.vpc_name
  }

  depends_on = [aws_internet_gateway.igw]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.app.id
  tags = {
    Name = "${var.vpc_name}-public"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route" "default-public" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.igw.id
}


# private
resource "aws_subnet" "private" {
  vpc_id     = aws_vpc.app.id
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "${var.vpc_name}-private"
  }
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.app.id
  tags = {
    Name = "${var.vpc_name}-private"
  }
}

resource "aws_route_table_association" "private" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.private.id
}

resource "aws_route" "default-private" {
  route_table_id         = aws_route_table.private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.ngw.id
}
