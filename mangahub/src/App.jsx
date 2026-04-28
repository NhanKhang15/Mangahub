import { useState, useMemo } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { Header, Footer } from './components/Layout';
import { Home } from './components/Home';
import { Library } from './components/Library';
import { MangaDetail } from './components/MangaDetail';
import { Login } from './components/Login';

export default function App() {
  const [currentView, setCurrentView] = useState('home');
  const [selectedManga, setSelectedManga] = useState(null);

  const handleSelectManga = (manga) => {
    setSelectedManga(manga);
    setCurrentView('detail');
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const navigate = (view) => {
    setCurrentView(view);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const renderView = () => {
    switch (currentView) {
      case 'home':
        return <Home onSelectManga={handleSelectManga} />;
      case 'library':
        return <Library onSelectManga={handleSelectManga} />;
      case 'detail':
        return selectedManga ? <MangaDetail manga={selectedManga} onBack={() => setCurrentView('home')} /> : null;
      case 'login':
        return <Login />;
      default:
        return <Home onSelectManga={handleSelectManga} />;
    }
  };

  return (
    <div className="min-h-screen flex flex-col">
      <Header currentView={currentView} onNavigate={navigate} />
      
      <main className="flex-1 max-w-7xl mx-auto w-full px-6 md:px-12 pt-28 pb-20">
        <AnimatePresence mode="wait">
          <motion.div
            key={currentView + (selectedManga?.id || '')}
            initial={{ opacity: 0, scale: 0.99 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 1.01 }}
            transition={{ duration: 0.3, ease: 'easeOut' }}
          >
            {renderView()}
          </motion.div>
        </AnimatePresence>
      </main>

      <Footer />
    </div>
  );
}
