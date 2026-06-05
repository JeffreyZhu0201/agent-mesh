import React, { useEffect, useRef } from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Container,
  Box,
  Grid,
  Card,
  CardContent,
  Chip,
  Avatar,
  IconButton,
} from '@mui/material';
import {
  AutoAwesome,
  Extension,
  Security,
  Speed,
  ArrowForward,
  Star,
  Menu as MenuIcon,
} from '@mui/icons-material';
import { styled } from '@mui/material/styles';
import { keyframes } from '@mui/system';

// Animation keyframes
const fadeInUp = keyframes`
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
`;

const float = keyframes`
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-10px); }
`;

const pulse = keyframes`
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
`;

// Styled components
const HeroSection = styled(Box)(({ theme }) => ({
  minHeight: '100vh',
  display: 'flex',
  alignItems: 'center',
  background: 'linear-gradient(135deg, #0F172A 0%, #1E293B 50%, #0F172A 100%)',
  position: 'relative',
  overflow: 'hidden',
  '&::before': {
    content: '""',
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    background: 'radial-gradient(circle at 20% 50%, rgba(34, 197, 94, 0.1) 0%, transparent 50%), radial-gradient(circle at 80% 50%, rgba(99, 102, 241, 0.1) 0%, transparent 50%)',
    pointerEvents: 'none',
  },
}));

const FeatureCardStyled = styled(Card)(({ theme }) => ({
  height: '100%',
  background: 'linear-gradient(145deg, #1E293B 0%, #0F172A 100%)',
  border: '1px solid #334155',
  borderRadius: 16,
  transition: 'all 0.3s ease-in-out',
  '&:hover': {
    transform: 'translateY(-8px)',
    borderColor: '#22C55E',
    boxShadow: '0 20px 40px rgba(34, 197, 94, 0.15)',
  },
}));

const FloatingOrb = styled(Box)({
  position: 'absolute',
  borderRadius: '50%',
  filter: 'blur(60px)',
  animation: `${float} 6s ease-in-out infinite`,
});

export function Home() {
  const featuresRef = useRef<HTMLDivElement>(null);
  const testimonialsRef = useRef<HTMLDivElement>(null);
  const ctaRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observerOptions = {
      threshold: 0.1,
      rootMargin: '0px 0px -50px 0px',
    };

    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('animate-in');
        }
      });
    }, observerOptions);

    [featuresRef.current, testimonialsRef.current, ctaRef.current].forEach((ref) => {
      if (ref) observer.observe(ref);
    });

    return () => observer.disconnect();
  }, []);

  const features = [
    {
      icon: <AutoAwesome sx={{ fontSize: 48, color: '#22C55E' }} />,
      title: 'AI Native',
      description: 'Built-in Eino integration for seamless AI agent orchestration and LLM management',
    },
    {
      icon: <Extension sx={{ fontSize: 48, color: '#6366F1' }} />,
      title: 'Plugin System',
      description: 'Dynamic plugin loading for unlimited extensibility. Build and deploy plugins at runtime',
    },
    {
      icon: <Security sx={{ fontSize: 48, color: '#F59E0B' }} />,
      title: 'SaaS Ready',
      description: 'Multi-tenant architecture with schema-level isolation for enterprise security',
    },
    {
      icon: <Speed sx={{ fontSize: 48, color: '#EC4899' }} />,
      title: 'High Performance',
      description: 'Go-zero microservices with gRPC communication for blazing fast responses',
    },
  ];

  const testimonials = [
    {
      name: 'Sarah Chen',
      role: 'CTO at TechCorp',
      content: 'AgentMesh accelerated our AI product development by 3x. The plugin system is incredibly well-designed.',
      rating: 5,
    },
    {
      name: 'Michael Park',
      role: 'Lead Developer at StartupXYZ',
      content: 'Finally, a framework that makes multi-tenant AI apps easy. The architecture is clean and extensible.',
      rating: 5,
    },
    {
      name: 'Emily Johnson',
      role: 'Product Manager at AI Labs',
      content: 'We built our entire SaaS platform on AgentMesh. The code quality and documentation are excellent.',
      rating: 5,
    },
  ];

  return (
    <Box sx={{ bgcolor: '#0F172A', color: '#F8FAFC', minHeight: '100vh' }}>
      {/* Navigation */}
      <AppBar
        position="fixed"
        sx={{
          bgcolor: 'rgba(15, 23, 42, 0.8)',
          backdropFilter: 'blur(10px)',
          borderBottom: '1px solid',
          borderColor: '#334155',
          boxShadow: 'none',
        }}
      >
        <Container maxWidth="lg">
          <Toolbar disableGutters sx={{ py: 1 }}>
            <Typography
              variant="h6"
              sx={{
                flexGrow: 1,
                fontWeight: 700,
                fontFamily: '"Space Grotesk", sans-serif',
                background: 'linear-gradient(90deg, #22C55E, #6366F1)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}
            >
              AgentMesh
            </Typography>
            <Button
              color="inherit"
              href="/app"
              sx={{
                color: '#F8FAFC',
                '&:hover': { bgcolor: 'rgba(255,255,255,0.1)' },
              }}
            >
              Login
            </Button>
            <Button
              variant="contained"
              sx={{
                ml: 2,
                bgcolor: '#22C55E',
                color: '#0F172A',
                fontWeight: 600,
                '&:hover': { bgcolor: '#16A34A' },
              }}
            >
              Get Started
            </Button>
          </Toolbar>
        </Container>
      </AppBar>

      {/* Hero Section */}
      <HeroSection>
        {/* Floating orbs */}
        <FloatingOrb
          sx={{
            width: 400,
            height: 400,
            bgcolor: 'rgba(34, 197, 94, 0.3)',
            top: '10%',
            left: '10%',
            animationDelay: '0s',
          }}
        />
        <FloatingOrb
          sx={{
            width: 300,
            height: 300,
            bgcolor: 'rgba(99, 102, 241, 0.3)',
            top: '20%',
            right: '15%',
            animationDelay: '2s',
          }}
        />
        <FloatingOrb
          sx={{
            width: 200,
            height: 200,
            bgcolor: 'rgba(236, 72, 153, 0.2)',
            bottom: '20%',
            left: '20%',
            animationDelay: '4s',
          }}
        />

        <Container maxWidth="lg" sx={{ position: 'relative', zIndex: 1 }}>
          <Grid container spacing={6} alignItems="center">
            <Grid item xs={12} md={6}>
              <Box
                sx={{
                  animation: `${fadeInUp} 0.8s ease-out`,
                }}
              >
                <Chip
                  label="Now in Beta"
                  sx={{
                    mb: 3,
                    bgcolor: 'rgba(34, 197, 94, 0.2)',
                    color: '#22C55E',
                    fontWeight: 600,
                    borderRadius: 2,
                    px: 1,
                  }}
                />
                <Typography
                  variant="h2"
                  sx={{
                    fontWeight: 700,
                    mb: 2,
                    fontFamily: '"Space Grotesk", sans-serif',
                    fontSize: { xs: '2.5rem', md: '3.5rem' },
                    lineHeight: 1.1,
                  }}
                >
                  Build AI Applications
                  <Box
                    component="span"
                    sx={{
                      display: 'block',
                      background: 'linear-gradient(90deg, #22C55E, #6366F1)',
                      WebkitBackgroundClip: 'text',
                      WebkitTextFillColor: 'transparent',
                    }}
                  >
                    10x Faster
                  </Box>
                </Typography>
                <Typography
                  variant="h6"
                  sx={{
                    mb: 4,
                    color: '#94A3B8',
                    fontWeight: 400,
                    fontFamily: '"DM Sans", sans-serif',
                    lineHeight: 1.6,
                  }}
                >
                  A comprehensive SaaS framework with go-zero microservices,
                  React frontend, and Eino-powered AI agents.
                </Typography>
                <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                  <Button
                    variant="contained"
                    size="large"
                    endIcon={<ArrowForward />}
                    sx={{
                      px: 4,
                      py: 1.5,
                      bgcolor: '#22C55E',
                      color: '#0F172A',
                      fontWeight: 600,
                      borderRadius: 2,
                      '&:hover': {
                        bgcolor: '#16A34A',
                        transform: 'translateY(-2px)',
                      },
                      transition: 'all 0.2s ease-in-out',
                    }}
                  >
                    Start Building
                  </Button>
                  <Button
                    variant="outlined"
                    size="large"
                    sx={{
                      px: 4,
                      py: 1.5,
                      color: '#F8FAFC',
                      borderColor: '#334155',
                      fontWeight: 600,
                      borderRadius: 2,
                      '&:hover': {
                        borderColor: '#6366F1',
                        bgcolor: 'rgba(99, 102, 241, 0.1)',
                      },
                      transition: 'all 0.2s ease-in-out',
                    }}
                  >
                    View Demo
                  </Button>
                </Box>
              </Box>
            </Grid>
            <Grid item xs={12} md={6}>
              <Box
                sx={{
                  animation: `${fadeInUp} 0.8s ease-out`,
                  animationDelay: '0.2s',
                  animationFillMode: 'both',
                }}
              >
                <Box
                  component="img"
                  src="/hero-illustration.svg"
                  alt="AgentMesh Architecture"
                  sx={{
                    width: '100%',
                    maxWidth: 500,
                    borderRadius: 4,
                    filter: 'drop-shadow(0 20px 60px rgba(34, 197, 94, 0.3))',
                  }}
                  onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                    const target = e.target as HTMLImageElement;
                    target.style.display = 'none';
                  }}
                />
              </Box>
            </Grid>
          </Grid>
        </Container>
      </HeroSection>

      {/* Features Section */}
      <Box
        ref={featuresRef}
        sx={{
          py: 16,
          bgcolor: '#0F172A',
          position: 'relative',
          '&.animate-in .feature-item': {
            animation: `${fadeInUp} 0.6s ease-out forwards`,
          },
        }}
      >
        <Container maxWidth="lg">
          <Box sx={{ textAlign: 'center', mb: 12 }}>
            <Typography
              variant="h3"
              sx={{
                fontWeight: 700,
                mb: 2,
                fontFamily: '"Space Grotesk", sans-serif',
              }}
            >
              Everything You Need
            </Typography>
            <Typography
              variant="h6"
              sx={{
                color: '#94A3B8',
                fontFamily: '"DM Sans", sans-serif',
                maxWidth: 600,
                mx: 'auto',
              }}
            >
              A complete framework for building production-ready AI SaaS applications
            </Typography>
          </Box>
          <Grid container spacing={4}>
            {features.map((feature, index) => (
              <Grid item xs={12} sm={6} md={3} key={index}>
                <Box
                  className="feature-item"
                  sx={{
                    opacity: 0,
                    animationDelay: `${index * 0.1}s`,
                  }}
                >
                  <FeatureCardStyled>
                    <CardContent sx={{ textAlign: 'center', p: 4 }}>
                      <Box sx={{ mb: 3, display: 'flex', justifyContent: 'center' }}>
                        {feature.icon}
                      </Box>
                      <Typography
                        variant="h6"
                        sx={{
                          mb: 2,
                          fontWeight: 600,
                          fontFamily: '"Space Grotesk", sans-serif',
                        }}
                      >
                        {feature.title}
                      </Typography>
                      <Typography
                        variant="body2"
                        sx={{ color: '#94A3B8', fontFamily: '"DM Sans", sans-serif' }}
                      >
                        {feature.description}
                      </Typography>
                    </CardContent>
                  </FeatureCardStyled>
                </Box>
              </Grid>
            ))}
          </Grid>
        </Container>
      </Box>

      {/* Testimonials */}
      <Box
        ref={testimonialsRef}
        sx={{
          py: 16,
          bgcolor: '#1E293B',
          '&.animate-in .testimonial-item': {
            animation: `${fadeInUp} 0.6s ease-out forwards`,
          },
        }}
      >
        <Container maxWidth="lg">
          <Typography
            variant="h3"
            align="center"
            sx={{
              mb: 12,
              fontWeight: 700,
              fontFamily: '"Space Grotesk", sans-serif',
            }}
          >
            Loved by Developers
          </Typography>
          <Grid container spacing={4}>
            {testimonials.map((testimonial, index) => (
              <Grid item xs={12} md={4} key={index}>
                <Box
                  className="testimonial-item"
                  sx={{ opacity: 0, animationDelay: `${index * 0.1}s` }}
                >
                  <Card
                    sx={{
                      height: '100%',
                      bgcolor: '#0F172A',
                      border: '1px solid #334155',
                      borderRadius: 3,
                      p: 2,
                    }}
                  >
                    <CardContent>
                      <Box sx={{ display: 'flex', mb: 2 }}>
                        {Array.from({ length: testimonial.rating }).map((_, i) => (
                          <Star
                            key={i}
                            sx={{ color: '#22C55E', fontSize: 20 }}
                          />
                        ))}
                      </Box>
                      <Typography
                        sx={{
                          mb: 3,
                          fontStyle: 'italic',
                          color: '#F8FAFC',
                          fontFamily: '"DM Sans", sans-serif',
                          lineHeight: 1.6,
                        }}
                      >
                        "{testimonial.content}"
                      </Typography>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                        <Avatar
                          sx={{
                            bgcolor: 'linear-gradient(135deg, #22C55E, #6366F1)',
                            width: 40,
                            height: 40,
                          }}
                        >
                          {testimonial.name[0]}
                        </Avatar>
                        <Box>
                          <Typography
                            variant="subtitle2"
                            sx={{ fontWeight: 600, color: '#F8FAFC' }}
                          >
                            {testimonial.name}
                          </Typography>
                          <Typography
                            variant="body2"
                            sx={{ color: '#94A3B8' }}
                          >
                            {testimonial.role}
                          </Typography>
                        </Box>
                      </Box>
                    </CardContent>
                  </Card>
                </Box>
              </Grid>
            ))}
          </Grid>
        </Container>
      </Box>

      {/* CTA Section */}
      <Box
        ref={ctaRef}
        sx={{
          py: 16,
          background: 'linear-gradient(135deg, #22C55E 0%, #6366F1 100%)',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <Container maxWidth="md" align="center">
          <Typography
            variant="h4"
            sx={{
              mb: 2,
              fontWeight: 700,
              fontFamily: '"Space Grotesk", sans-serif',
              color: '#FFFFFF',
            }}
          >
            Ready to Build?
          </Typography>
          <Typography
            variant="h6"
            sx={{
              mb: 4,
              color: 'rgba(255, 255, 255, 0.9)',
              fontFamily: '"DM Sans", sans-serif',
            }}
          >
            Start your AI application development today
          </Typography>
          <Button
            variant="contained"
            size="large"
            endIcon={<ArrowForward />}
            sx={{
              px: 6,
              py: 2,
              bgcolor: '#0F172A',
              color: '#FFFFFF',
              fontWeight: 600,
              borderRadius: 2,
              '&:hover': {
                bgcolor: '#1E293B',
                transform: 'translateY(-2px)',
              },
              transition: 'all 0.2s ease-in-out',
            }}
          >
            Get Started Free
          </Button>
        </Container>
      </Box>

      {/* Footer */}
      <Box
        sx={{
          py: 4,
          bgcolor: '#0F172A',
          borderTop: '1px solid',
          borderColor: '#334155',
        }}
      >
        <Container maxWidth="lg">
          <Typography align="center" sx={{ color: '#64748B' }}>
            © 2024 AgentMesh. All rights reserved.
          </Typography>
        </Container>
      </Box>

      {/* Global styles */}
      <style>
        {`
          @import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;700&family=Space+Grotesk:wght@400;500;600;700&display=swap');

          * {
            box-sizing: border-box;
          }

          html {
            scroll-behavior: smooth;
          }

          @media (prefers-reduced-motion: reduce) {
            * {
              animation-duration: 0.01ms !important;
              animation-iteration-count: 1 !important;
              transition-duration: 0.01ms !important;
            }
          }
        `}
      </style>
    </Box>
  );
}
