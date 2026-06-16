export interface Dish {
  id: string;
  name: string;
  description?: string;
  price: number;
  image: string;
  category: 'main' | 'vegetables' | 'alcohol' | 'drinks';
  categoryName?: string;
  calories?: number;
  weight?: number;
}

export const MAIN_DISHES: Dish[] = [
  {
    id: '1',
    name: 'Нью-Йорк стрип или стейк в полоску',
    price: 650,
    image: '/src/assets/images/hero_steak_plate_1779197902033.png',
    category: 'main',
  },
  {
    id: '2',
    name: 'Стейк из лосося с фирменным соусом',
    price: 920,
    image: '/src/assets/images/salmon_steak_plate_1779197942330.png',
    category: 'main',
  },
  {
    id: '3',
    name: 'Говяжья вырезка с овощами и нашим ягодным коктейлем',
    price: 1150,
    image: '/src/assets/images/hero_steak_plate_1779197902033.png',
    category: 'main',
  },
  {
    id: '4',
    name: 'Судак с ягодами и лемоном',
    price: 410,
    image: '/src/assets/images/grilled_fish_plate_1779197923227.png',
    category: 'main',
  },
];

export const VEGETABLES: Dish[] = [
  {
    id: '5',
    name: 'Овощи гриль в медово-горчичном соусе',
    price: 450,
    image: '/src/assets/images/hero_steak_plate_1779197902033.png',
    category: 'vegetables',
  },
  {
    id: '6',
    name: 'Летний салат с нерафинированным маслом',
    price: 115,
    image: '/src/assets/images/fresh_salad_plate_1779198150233.png',
    category: 'vegetables',
  },
  {
    id: '7',
    name: 'Классический «Цезарь»',
    price: 149,
    image: '/src/assets/images/fresh_salad_plate_1779198150233.png',
    category: 'vegetables',
  },
  {
    id: '8',
    name: 'Белые грибы в сметане с зеленью',
    price: 650,
    image: '/src/assets/images/creamy_mushrooms_plate_1779198170202.png',
    category: 'vegetables',
  },
];
